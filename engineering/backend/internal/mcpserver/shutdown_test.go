package mcpserver

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeUntilDone(t *testing.T) {
	tests := []struct {
		name         string
		sendRequest  bool          // start a request before shutdown begins
		handlerDelay time.Duration // how long that request takes
		grace        time.Duration
		wantErr      bool // serveUntilDone reports in-flight requests were cut off
		wantFinished bool // the in-flight request got its response
	}{
		{name: "idle server stops cleanly", grace: time.Second},
		{name: "in-flight request finishes within the grace period", sendRequest: true, handlerDelay: 200 * time.Millisecond, grace: 5 * time.Second, wantFinished: true},
		{name: "in-flight request outlasts the grace period", sendRequest: true, handlerDelay: 10 * time.Second, grace: 100 * time.Millisecond, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			started := make(chan struct{})
			srv := &http.Server{
				ReadHeaderTimeout: time.Second,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					close(started)
					select {
					case <-time.After(tt.handlerDelay):
						w.WriteHeader(http.StatusOK)
					case <-r.Context().Done(): // connection closed by a forced shutdown
					}
				}),
			}
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- serveUntilDone(ctx, srv, ln, tt.grace) }()

			finished := make(chan bool, 1)
			if tt.sendRequest {
				go func() {
					resp, err := http.Get("http://" + ln.Addr().String())
					if err != nil {
						finished <- false
						return
					}
					resp.Body.Close()
					finished <- resp.StatusCode == http.StatusOK
				}()
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					t.Fatal("request never reached the handler")
				}
			}

			cancel() // what SIGTERM does

			select {
			case err := <-done:
				if (err != nil) != tt.wantErr {
					t.Errorf("serveUntilDone error = %v, wantErr %v", err, tt.wantErr)
				}
			case <-time.After(15 * time.Second):
				t.Fatal("serveUntilDone did not return")
			}

			if tt.sendRequest {
				select {
				case got := <-finished:
					if got != tt.wantFinished {
						t.Errorf("in-flight request finished = %v, want %v", got, tt.wantFinished)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("client never got an outcome for the in-flight request")
				}
			}
		})
	}
}
