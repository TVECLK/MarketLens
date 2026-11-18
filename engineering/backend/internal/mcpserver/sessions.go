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
//
// A session is also "claimed" when shutdown decides to close it. Deciding that
// the session is idle and closing it must not be separate steps: a POST that
// slipped in between would find its session gone. So the decision is made
// under the same lock that admits POSTs (closeIfIdle), and once a session is
// claimed, later POSTs for it are refused with a retryable response, while
// those already admitted finish normally.
type sessionTracker struct {
	mu       sync.Mutex
	inflight map[string]int
	claimed  map[string]bool // sessions being closed: no new POST is admitted
}

func newSessionTracker() *sessionTracker {
	return &sessionTracker{inflight: make(map[string]int), claimed: make(map[string]bool)}
}

// track wraps next, counting each session's in-flight POST requests. A POST
// for a session that is being closed is refused with 503 and Retry-After, so
// the client retries instead of finding the session gone halfway through.
func (t *sessionTracker) track(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get(sessionIDHeader); id != "" && r.Method == http.MethodPost {
			if !t.begin(id) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "the server is shutting down, retry shortly", http.StatusServiceUnavailable)
				return
			}
			defer t.end(id)
		}
		next.ServeHTTP(w, r)
	})
}

// begin admits a POST for the session, or reports false if the session is
// being closed.
func (t *sessionTracker) begin(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.claimed[id] {
		return false
	}
	t.inflight[id]++
	return true
}

func (t *sessionTracker) end(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inflight[id]--; t.inflight[id] <= 0 {
		delete(t.inflight, id)
	}
}

// closeIfIdle claims the session for closing if it has no POST in flight, and
// reports whether it did. Checking and claiming happen under one lock, so no
// POST can be admitted in between; from then on begin refuses the session. It
// returns false if a POST is in flight or the session was already claimed, so
// a session is claimed once. The caller closes the session after this returns,
// without the lock held.
func (t *sessionTracker) closeIfIdle(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inflight[id] > 0 || t.claimed[id] {
		return false
	}
	t.claimed[id] = true
	return true
}

// release gives up a claim, for a session whose close failed, so POSTs are
// admitted again and the close can be retried.
func (t *sessionTracker) release(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.claimed, id)
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
			id := ss.ID()
			if !tracker.closeIfIdle(id) {
				continue
			}
			// Each close runs on its own goroutine. Close waits for any tool
			// still running on the session (one whose client disconnected is no
			// longer counted as in flight, but keeps running), so closing in
			// this loop would let one stuck session hold up every other idle
			// stream. The claim taken above stops later polls starting the same
			// close again; a failed close releases it so the next poll retries.
			go func() {
				if err := ss.Close(); err != nil {
					log.Printf("closing idle MCP session %s: %v", id, err)
					tracker.release(id)
				}
			}()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
