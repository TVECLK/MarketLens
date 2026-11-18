package mcpserver

import (
	"context"
	"net"
	"net/http"
	"sync"
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
			serverStopped := make(chan struct{})
			go func() {
				defer close(serverStopped)
				done <- serveUntilDone(shutdownCtx, srv, ln, tt.grace, onShutdown)
			}()

			// stopServer signals shutdown and waits for the server goroutine,
			// with a bound: serveUntilDone gives up after the grace period, and
			// if it somehow does not, closing the listener ends it. It is safe
			// to call more than once, and it runs even if the test fails early.
			var stopOnce sync.Once
			stopServer := func() {
				stopOnce.Do(func() {
					beginShutdown()
					select {
					case <-serverStopped:
					case <-time.After(tt.grace + 2*time.Second):
						srv.Close()
						select {
						case <-serverStopped:
						case <-time.After(2 * time.Second):
							t.Error("the server goroutine did not stop")
						}
					}
				})
			}
			t.Cleanup(stopServer)

			// Every client call uses this context, so cleanup can cancel a call
			// that is still pending when the test fails.
			callCtx, cancelCalls := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelCalls()

			client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
			session, err := client.Connect(callCtx, &mcp.StreamableClientTransport{Endpoint: "http://" + ln.Addr().String()}, nil)
			if err != nil {
				t.Fatalf("connect: %v", err)
			}
			// Runs before the cleanup above: cancel pending calls, stop the
			// server and wait for it, and only then close the session, with a
			// bound, because Close waits for calls that are still pending.
			t.Cleanup(func() {
				cancelCalls()
				stopServer()
				closed := make(chan struct{})
				go func() {
					session.Close()
					close(closed)
				}()
				select {
				case <-closed:
				case <-time.After(3 * time.Second):
					t.Log("closing the client session did not finish within 3s")
				}
			})

			// The client opens its standalone stream just after initializing.
			deadline := time.Now().Add(3 * time.Second)
			for openStreams.Load() == 0 {
				if time.Now().After(deadline) {
					t.Fatal("the client never opened its standalone SSE stream, so there is nothing to test")
				}
				time.Sleep(10 * time.Millisecond)
			}

			type callOutcome struct {
				res *mcp.CallToolResult
				err error
			}
			callDone := make(chan callOutcome, 1)
			if tt.toolDelay > 0 {
				go func() {
					res, err := session.CallTool(callCtx, &mcp.CallToolParams{Name: "slow"})
					callDone <- callOutcome{res: res, err: err}
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
				case outcome := <-callDone:
					// CallTool returns a nil error for a tool that failed, with
					// IsError set on the result, so the result is checked too.
					switch {
					case outcome.err != nil:
						t.Errorf("the in-flight tool call failed: %v", outcome.err)
					case outcome.res.IsError:
						var text []string
						for _, c := range outcome.res.Content {
							if tc, ok := c.(*mcp.TextContent); ok {
								text = append(text, tc.Text)
							}
						}
						t.Errorf("the in-flight tool call returned an error result: %q", text)
					default:
						if !boolField(t, outcome.res, "done") {
							t.Error("the in-flight tool call did not report done: true")
						}
					}
				case <-time.After(3 * time.Second):
					t.Error("the in-flight tool call never returned")
				}
			}
		})
	}
}
