package mcpserver

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestShutdown_WithOpenSSEStream runs the real SDK handler behind the session
// tracker, connects a client whose standalone SSE stream stays open, and
// shuts the server down. Shutdown must not wait for the stream to end by
// itself, because it never does while the session lives.
func TestShutdown_WithOpenSSEStream(t *testing.T) {
	tests := []struct {
		name        string
		closeIdle   bool          // run closeIdleSessions during shutdown, as StartHTTP does
		toolDelay   time.Duration // run a tool call this long, and begin shutdown while it runs; 0 means no call
		grace       time.Duration
		wantErr     bool
		wantWithin  time.Duration // shutdown must finish within this long
		wantCallOK  bool          // the in-flight call must get its result
		minDuration time.Duration // shutdown must take at least this long
	}{
		{
			name:       "idle session: shutdown does not wait for the SSE stream",
			closeIdle:  true,
			grace:      10 * time.Second,
			wantWithin: 3 * time.Second,
		},
		{
			name:        "tool call running: shutdown waits for it, then ends the stream",
			closeIdle:   true,
			toolDelay:   500 * time.Millisecond,
			grace:       10 * time.Second,
			wantWithin:  5 * time.Second,
			wantCallOK:  true,
			minDuration: 200 * time.Millisecond,
		},
		{
			// Control: without closing idle sessions the open stream holds
			// Shutdown until the grace period ends. This is the behavior the
			// fix removes, and shows the test can tell the two apart.
			name:       "without closing idle sessions the open stream outlasts the grace period",
			closeIdle:  false,
			grace:      500 * time.Millisecond,
			wantErr:    true,
			wantWithin: 3 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
			toolStarted := make(chan struct{}, 1)
			mcp.AddTool(server, &mcp.Tool{Name: "slow"},
				func(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
					select {
					case toolStarted <- struct{}{}:
					default:
					}
					time.Sleep(tt.toolDelay)
					return nil, map[string]any{"done": true}, nil
				})

			tracker := newSessionTracker()
			sdkHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
			tracked := tracker.track(sdkHandler)
			var openStreams atomic.Int32
			srv := &http.Server{
				ReadHeaderTimeout: time.Second,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodGet {
						openStreams.Add(1)
						defer openStreams.Add(-1)
					}
					tracked.ServeHTTP(w, r)
				}),
			}
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}

			shutdownCtx, beginShutdown := context.WithCancel(context.Background())
			defer beginShutdown()
			var onShutdown func(context.Context)
			if tt.closeIdle {
				onShutdown = func(drainCtx context.Context) { closeIdleSessions(drainCtx, server, tracker) }
			}
			done := make(chan error, 1)
			go func() { done <- serveUntilDone(shutdownCtx, srv, ln, tt.grace, onShutdown) }()

			client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
			session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: "http://" + ln.Addr().String()}, nil)
			if err != nil {
				t.Fatalf("connect: %v", err)
			}
			defer session.Close()

			// The client opens its standalone stream just after initializing.
			deadline := time.Now().Add(3 * time.Second)
			for openStreams.Load() == 0 {
				if time.Now().After(deadline) {
					t.Fatal("the client never opened its standalone SSE stream, so there is nothing to test")
				}
				time.Sleep(10 * time.Millisecond)
			}

			callDone := make(chan error, 1)
			if tt.toolDelay > 0 {
				go func() {
					_, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "slow"})
					callDone <- err
				}()
				select {
				case <-toolStarted:
				case <-time.After(3 * time.Second):
					t.Fatal("the tool call never started")
				}
			}

			begun := time.Now()
			beginShutdown()

			select {
			case err := <-done:
				elapsed := time.Since(begun)
				if (err != nil) != tt.wantErr {
					t.Errorf("serveUntilDone error = %v, wantErr %v (after %v)", err, tt.wantErr, elapsed)
				}
				if elapsed > tt.wantWithin {
					t.Errorf("shutdown took %v, want within %v", elapsed, tt.wantWithin)
				}
				if elapsed < tt.minDuration {
					t.Errorf("shutdown took %v, want at least %v (it should wait for the running call)", elapsed, tt.minDuration)
				}
			case <-time.After(tt.wantWithin + 5*time.Second):
				t.Fatal("serveUntilDone did not return")
			}

			if tt.wantCallOK {
				select {
				case err := <-callDone:
					if err != nil {
						t.Errorf("the in-flight tool call failed: %v", err)
					}
				case <-time.After(3 * time.Second):
					t.Error("the in-flight tool call never returned")
				}
			}
		})
	}
}
