package mcpserver

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	sessionIDHeader = "Mcp-Session-Id"

	// sessionDrainPollInterval is how often closeIdleSessions looks for
	// sessions whose last request has finished.
	sessionDrainPollInterval = 50 * time.Millisecond
)

// sessionTracker counts the POST requests in flight for each MCP session. A
// POST carries a tool call and lasts until its response is written. The
// standalone SSE GET a client keeps open is deliberately not counted: it
// lasts as long as the session, and it is exactly what shutdown must not wait
// for.
type sessionTracker struct {
	mu       sync.Mutex
	inflight map[string]int
}

func newSessionTracker() *sessionTracker {
	return &sessionTracker{inflight: make(map[string]int)}
}

// track wraps next, counting each session's in-flight POST requests.
func (t *sessionTracker) track(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get(sessionIDHeader); id != "" && r.Method == http.MethodPost {
			t.begin(id)
			defer t.end(id)
		}
		next.ServeHTTP(w, r)
	})
}

func (t *sessionTracker) begin(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.inflight[id]++
}

func (t *sessionTracker) end(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inflight[id]--; t.inflight[id] <= 0 {
		delete(t.inflight, id)
	}
}

// idle reports whether the session has no POST request in flight.
func (t *sessionTracker) idle(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.inflight[id] == 0
}

// closeIdleSessions closes every MCP session as soon as it has no request in
// flight, until none are left or ctx ends. Closing a session ends its
// standalone SSE stream. Without this, http.Server.Shutdown waits on that
// stream, which stays open for the whole session, so every shutdown would run
// into the grace period even with no tool call running. A session with a call
// in progress is left alone until that call's response has been written.
func closeIdleSessions(ctx context.Context, server *mcp.Server, tracker *sessionTracker) {
	ticker := time.NewTicker(sessionDrainPollInterval)
	defer ticker.Stop()
	for {
		var sessions []*mcp.ServerSession
		for ss := range server.Sessions() {
			sessions = append(sessions, ss)
		}
		if len(sessions) == 0 {
			return
		}
		for _, ss := range sessions {
			if tracker.idle(ss.ID()) {
				if err := ss.Close(); err != nil {
					log.Printf("closing idle MCP session %s: %v", ss.ID(), err)
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
