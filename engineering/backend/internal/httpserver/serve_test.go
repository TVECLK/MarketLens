package httpserver

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
		wantErr      bool // ServeUntilDone reports in-flight requests were cut off
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
			go func() { done <- ServeUntilDone(ctx, srv, ln, tt.grace, nil) }()

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
					t.Errorf("ServeUntilDone error = %v, wantErr %v", err, tt.wantErr)
				}
			case <-time.After(15 * time.Second):
				t.Fatal("ServeUntilDone did not return")
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

// TestNewServer_Timeouts checks every stage of a request has a timeout. A zero
// value means no limit, so a regression back to a bare server, which is what
// gosec flags and what lets a slow client hold a connection open, fails here.
func TestNewServer_Timeouts(t *testing.T) {
	srv := NewServer(http.NotFoundHandler())

	tests := []struct {
		name string
		got  time.Duration
	}{
		{"ReadHeaderTimeout", srv.ReadHeaderTimeout},
		{"ReadTimeout", srv.ReadTimeout},
		{"WriteTimeout", srv.WriteTimeout},
		{"IdleTimeout", srv.IdleTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got <= 0 {
				t.Errorf("%s = %v, want a positive timeout", tt.name, tt.got)
			}
		})
	}
	if srv.ReadHeaderTimeout != ReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", srv.ReadHeaderTimeout, ReadHeaderTimeout)
	}
	if srv.Handler == nil {
		t.Error("server has no handler")
	}
}

// TestNewServer_ClosesSlowClient does what a Slowloris client does: opens a
// connection and sends only part of a request. The server must hang up on it
// once the header timeout passes, not wait for the rest.
func TestNewServer_ClosesSlowClient(t *testing.T) {
	srv := NewServer(http.NotFoundHandler())
	srv.ReadHeaderTimeout = 200 * time.Millisecond // shortened so the test is quick

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- ServeUntilDone(ctx, srv, ln, time.Second, nil) }()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	// Part of a request, never finished: no blank line ends the headers.
	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: example\r\nX-Slow: ")); err != nil {
		t.Fatal(err)
	}

	// The server closing the connection shows up as a read that returns, with
	// either an error response or EOF, well before this deadline.
	start := time.Now()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 512)
	_, readErr := conn.Read(buf)
	elapsed := time.Since(start)

	if elapsed >= 4*time.Second {
		t.Errorf("the server held the slow connection for %v (read error: %v), want it closed shortly after the header timeout", elapsed, readErr)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ServeUntilDone did not return")
	}
}
