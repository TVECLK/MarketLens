package mcpserver

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"marketlens-go-backend/internal/httpserver"
)

func TestSessionTracker_Claim(t *testing.T) {
	type step struct {
		op   string // "begin", "end", "claim" or "release"
		id   string
		want bool // begin: admitted; claim: claimed; ignored for end and release
	}
	tests := []struct {
		name  string
		steps []step
	}{
		{"an idle session can be claimed", []step{{"claim", "s1", true}}},
		{"a claimed session refuses new POSTs", []step{{"claim", "s1", true}, {"begin", "s1", false}}},
		{"a session with an admitted POST is not claimed", []step{
			{"begin", "s1", true}, {"claim", "s1", false}, {"end", "s1", false}, {"claim", "s1", true},
		}},
		{"a session is claimed only once", []step{{"claim", "s1", true}, {"claim", "s1", false}}},
		{"a claim does not affect other sessions", []step{{"claim", "s1", true}, {"begin", "s2", true}}},
		{"a released claim admits POSTs again", []step{{"claim", "s1", true}, {"release", "s1", false}, {"begin", "s1", true}}},
		{"POSTs admitted before the claim attempt keep the session open until all finish", []step{
			{"begin", "s1", true}, {"begin", "s1", true},
			{"end", "s1", false}, {"claim", "s1", false},
			{"end", "s1", false}, {"claim", "s1", true},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newSessionTracker()
			for i, s := range tt.steps {
				switch s.op {
				case "begin":
					if got := tracker.begin(s.id); got != s.want {
						t.Fatalf("step %d: begin(%s) = %v, want %v", i, s.id, got, s.want)
					}
				case "claim":
					if got := tracker.closeIfIdle(s.id); got != s.want {
						t.Fatalf("step %d: closeIfIdle(%s) = %v, want %v", i, s.id, got, s.want)
					}
				case "end":
					tracker.end(s.id)
				case "release":
					tracker.release(s.id)
				}
			}
		})
	}
}

func TestSessionTracker_Track(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		sessionID      string
		wantStatus     int
		wantRetryAfter string
	}{
		{"POST for a session being closed is refused, retryably", http.MethodPost, "closing", http.StatusServiceUnavailable, "1"},
		{"POST for another session is admitted", http.MethodPost, "other", http.StatusOK, ""},
		{"GET stream of a session being closed is not refused", http.MethodGet, "closing", http.StatusOK, ""},
		{"POST without a session id (initialize) is admitted", http.MethodPost, "", http.StatusOK, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newSessionTracker()
			if !tracker.closeIfIdle("closing") {
				t.Fatal("could not claim the session for the test")
			}
			handler := tracker.track(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/mcp", nil)
			if tt.sessionID != "" {
				req.Header.Set(sessionIDHeader, tt.sessionID)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if got := w.Header().Get("Retry-After"); got != tt.wantRetryAfter {
				t.Errorf("Retry-After = %q, want %q", got, tt.wantRetryAfter)
			}
			// A request, admitted or refused, must not leave a count behind.
			if len(tracker.inflight) != 0 {
				t.Errorf("in-flight counts left behind: %v", tracker.inflight)
			}
		})
	}
}

// TestSessionTracker_ClaimIsAtomicWithAdmission races a POST being admitted
// against the decision to close its session. Exactly one of them may win: if
// the POST was admitted the session must not be claimed, and if the session
// was claimed the POST must be refused. Both winning is the bug where the
// session is closed under a POST that already got in. The race detector also
// watches the tracker's state.
func TestSessionTracker_ClaimIsAtomicWithAdmission(t *testing.T) {
	for i := 0; i < 2000; i++ {
		tracker := newSessionTracker()
		start := make(chan struct{})
		var wg sync.WaitGroup
		var admitted, claimed bool

		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			admitted = tracker.begin("s")
		}()
		go func() {
			defer wg.Done()
			<-start
			claimed = tracker.closeIfIdle("s")
		}()
		close(start)
		wg.Wait()

		if admitted == claimed {
			t.Fatalf("iteration %d: admitted=%v claimed=%v, want exactly one", i, admitted, claimed)
		}
	}
}

// TestCloseIdleSessions_SlowCloseDoesNotBlockOthers has one session whose
// tool keeps running after its client gave up, so closing it waits for the
// tool, and several idle sessions. The idle ones must close promptly all the
// same: if closes ran one after another, whichever came after the stuck one
// would wait for it, and HTTP shutdown would wait on their SSE streams.
func TestCloseIdleSessions_SlowCloseDoesNotBlockOthers(t *testing.T) {
	const (
		idleSessions = 4
		toolDuration = 2 * time.Second
		closeWithin  = time.Second // well under toolDuration
	)

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	toolStarted := make(chan struct{}, 1)
	mcp.AddTool(server, &mcp.Tool{Name: "stuck"},
		func(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
			toolStarted <- struct{}{}
			time.Sleep(toolDuration) // ignores its context, like a tool stuck on a database call
			return nil, map[string]any{"done": true}, nil
		})

	tracker := newSessionTracker()
	sdkHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	ts := httptest.NewServer(tracker.track(sdkHandler))
	t.Cleanup(ts.Close) // registered first, so it runs after the clients are closed

	connect := func() *mcp.ClientSession {
		t.Helper()
		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
		session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
		if err != nil {
			t.Fatalf("connect: %v", err)
		}
		t.Cleanup(func() { session.Close() })
		return session
	}

	// The stuck session: its client calls the tool and then gives up. The POST
	// ends, so the tracker no longer counts it, but the tool keeps running.
	stuck := connect()
	stuckID := stuck.ID()
	callCtx, cancelCall := context.WithCancel(context.Background())
	defer cancelCall()
	go stuck.CallTool(callCtx, &mcp.CallToolParams{Name: "stuck"})
	select {
	case <-toolStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("the tool call never started")
	}
	cancelCall()
	time.Sleep(300 * time.Millisecond)

	tracker.mu.Lock()
	stillCounted := tracker.inflight[stuckID]
	tracker.mu.Unlock()
	if stillCounted != 0 {
		t.Fatalf("setup: the abandoned POST is still counted as in flight (%d), so the session is not idle", stillCounted)
	}

	for i := 0; i < idleSessions; i++ {
		connect()
	}

	closeCtx, cancelClose := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClose()
	go closeIdleSessions(closeCtx, server, tracker)

	// Every idle session must be gone within closeWithin, although closing the
	// stuck one is still waiting for its tool.
	otherSessions := func() int {
		n := 0
		for ss := range server.Sessions() {
			if ss.ID() != stuckID {
				n++
			}
		}
		return n
	}
	deadline := time.Now().Add(closeWithin)
	for otherSessions() > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("%d idle sessions were still open after %v: closing the stuck session held them up", otherSessions(), closeWithin)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// The stuck session was claimed too, so the test really did have a close
	// that was waiting on the tool, not a session that was skipped.
	tracker.mu.Lock()
	stuckClaimed := tracker.claimed[stuckID]
	tracker.mu.Unlock()
	if !stuckClaimed {
		t.Error("the stuck session was never claimed for closing, so the scenario was not exercised")
	}
}

func TestWaitForCloses(t *testing.T) {
	tests := []struct {
		name       string
		closeAfter time.Duration // the pending close finishes this long after the wait starts; 0 means never
		ctxTimeout time.Duration // the drain deadline
		minElapsed time.Duration
		maxElapsed time.Duration
	}{
		{"no pending close: returns at once", 0, 2 * time.Second, 0, 300 * time.Millisecond}, // wg left empty below
		{"waits until the pending close has finished", 300 * time.Millisecond, 5 * time.Second, 250 * time.Millisecond, 2 * time.Second},
		{"gives up at the deadline when a close never finishes", 0, 300 * time.Millisecond, 250 * time.Millisecond, 2 * time.Second},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			release := make(chan struct{})
			defer close(release) // lets a close that "never finishes" end with the test
			switch {
			case i == 1: // a close that finishes after closeAfter
				wg.Add(1)
				time.AfterFunc(tt.closeAfter, wg.Done)
			case i == 2: // a close that never finishes on its own
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-release
				}()
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel()
			start := time.Now()
			waitForCloses(ctx, &wg)
			elapsed := time.Since(start)

			if elapsed < tt.minElapsed || elapsed > tt.maxElapsed {
				t.Errorf("waited %v, want between %v and %v", elapsed, tt.minElapsed, tt.maxElapsed)
			}
		})
	}
}

// TestShutdown_WaitsForToolWhoseClientDisconnected is the case where shutdown
// has no HTTP connection left to wait on. The client has no standalone SSE
// stream, calls a tool and disconnects; the tool keeps running on the server.
// Server.Shutdown alone would return at once, and the process could exit under
// the tool before it finalised its crawler run, so shutdown must wait for it.
func TestShutdown_WaitsForToolWhoseClientDisconnected(t *testing.T) {
	const toolDuration = 1200 * time.Millisecond

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	toolStarted := make(chan struct{}, 1)
	var toolFinished atomic.Bool
	mcp.AddTool(server, &mcp.Tool{Name: "finalise"},
		func(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
			toolStarted <- struct{}{}
			time.Sleep(toolDuration) // ignores its context, like a tool still closing its run
			toolFinished.Store(true)
			return nil, map[string]any{"done": true}, nil
		})

	tracker := newSessionTracker()
	sdkHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	srv := &http.Server{ReadHeaderTimeout: time.Second, Handler: tracker.track(sdkHandler)}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	shutdownCtx, beginShutdown := context.WithCancel(context.Background())
	defer beginShutdown()
	done := make(chan error, 1)
	go func() {
		done <- httpserver.ServeUntilDone(shutdownCtx, srv, ln, 10*time.Second, func(drainCtx context.Context) {
			closeIdleSessions(drainCtx, server, tracker)
		})
	}()

	// No standalone SSE stream: once the call is abandoned, nothing keeps an
	// HTTP connection open for Shutdown to wait on.
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:             "http://" + ln.Addr().String(),
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		closed := make(chan struct{})
		go func() {
			session.Close()
			close(closed)
		}()
		select {
		case <-closed:
		case <-time.After(3 * time.Second):
		}
	})

	callCtx, cancelCall := context.WithCancel(context.Background())
	defer cancelCall()
	go session.CallTool(callCtx, &mcp.CallToolParams{Name: "finalise"})
	select {
	case <-toolStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("the tool call never started")
	}
	cancelCall() // the client disconnects; the tool carries on
	time.Sleep(300 * time.Millisecond)

	tracker.mu.Lock()
	counted := len(tracker.inflight)
	tracker.mu.Unlock()
	if counted != 0 || toolFinished.Load() {
		t.Fatalf("setup: want the POST gone and the tool still running (counted=%d, finished=%v)", counted, toolFinished.Load())
	}

	begun := time.Now()
	beginShutdown()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("ServeUntilDone error = %v, want nil", err)
		}
		if !toolFinished.Load() {
			t.Error("ServeUntilDone returned before the tool had finished: the process would exit under it")
		}
		if elapsed := time.Since(begun); elapsed < 500*time.Millisecond {
			t.Errorf("returned after %v, want it to wait out the tool's remaining time", elapsed)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("ServeUntilDone did not return")
	}
}

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
				done <- httpserver.ServeUntilDone(shutdownCtx, srv, ln, tt.grace, onShutdown)
			}()

			// stopServer signals shutdown and waits for the server goroutine,
			// with a bound: ServeUntilDone gives up after the grace period, and
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
					t.Errorf("ServeUntilDone error = %v, wantErr %v (after %v)", err, tt.wantErr, elapsed)
				}
				if elapsed > tt.wantWithin {
					t.Errorf("shutdown took %v, want within %v", elapsed, tt.wantWithin)
				}
				if elapsed < tt.minDuration {
					t.Errorf("shutdown took %v, want at least %v (it should wait for the running call)", elapsed, tt.minDuration)
				}
			case <-time.After(tt.wantWithin + 5*time.Second):
				t.Fatal("ServeUntilDone did not return")
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
