package api

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"marketlens-go-backend/internal/httpserver"
)

// TestLiftWriteDeadline runs a route that answers later than the server's write
// timeout, on a real server (the deadline only exists on a real connection),
// with and without liftWriteDeadline. Without it the client never gets the
// response, although the handler finished; that is the batch-save failure.
func TestLiftWriteDeadline(t *testing.T) {
	const (
		writeTimeout = 300 * time.Millisecond // shortened from the real 60s so the test is quick
		handlerTime  = 800 * time.Millisecond // longer than writeTimeout
	)

	gin.SetMode(gin.TestMode)
	// The handler also reports whether its request context survived: batch-save
	// passes that context to ingestion, so a context cancelled at the read
	// timeout would stop ingestion partway, whatever the write deadline is.
	slow := func(c *gin.Context) {
		time.Sleep(handlerTime)
		if c.Request.Context().Err() != nil {
			c.String(http.StatusOK, "request context cancelled")
			return
		}
		c.String(http.StatusOK, "done")
	}
	r := gin.New()
	r.GET("/normal", slow)
	r.GET("/lifted", liftWriteDeadline(), slow)

	srv := httpserver.NewServer(r.Handler())
	srv.WriteTimeout = writeTimeout
	srv.ReadTimeout = writeTimeout // shortened too: it must not cancel the request while the handler runs
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- httpserver.ServeUntilDone(ctx, srv, ln, 5*time.Second, nil) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("the server did not stop")
		}
	})

	tests := []struct {
		name         string
		path         string
		wantResponse bool // the client receives the 200 and its body
	}{
		{"a normal route loses a response written after the write timeout", "/normal", false},
		{"a route with the lifted deadline still delivers its response", "/lifted", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get("http://" + ln.Addr().String() + tt.path)
			if !tt.wantResponse {
				if err == nil {
					resp.Body.Close()
					t.Fatalf("got status %d, want the connection to be cut at the write deadline", resp.StatusCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("the response was lost: %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK || string(body) != "done" {
				t.Errorf("got %d %q, want 200 %q", resp.StatusCode, body, "done")
			}
		})
	}
}

// TestLiftWriteDeadline_UnsupportedWriter checks that a writer that cannot
// change its deadline (a test recorder, or a connection behind a proxy that
// hides it) does not break the request: the handler still runs.
func TestLiftWriteDeadline_UnsupportedWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", liftWriteDeadline(), func(c *gin.Context) { c.String(http.StatusOK, "ran") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if w.Code != http.StatusOK || w.Body.String() != "ran" {
		t.Errorf("got %d %q, want 200 %q", w.Code, w.Body.String(), "ran")
	}
}
