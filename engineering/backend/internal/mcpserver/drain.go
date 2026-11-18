package mcpserver

import (
	"context"
	"time"
)

const (
	// drainFinalizeMargin is the part of shutdownGracePeriod kept back, after
	// shutdown begins, to close the crawler run and answer the client.
	drainFinalizeMargin = 5 * time.Second

	// drainWorkBudget is how long ingestion may keep running once shutdown
	// begins. Ingestion has no bound of its own (each job makes about 11
	// sequential LLM calls, and retries add more), so without this budget a
	// submission can outlast shutdownGracePeriod and be killed before its run
	// is closed.
	drainWorkBudget = shutdownGracePeriod - drainFinalizeMargin
)

// drainBudget is one process-wide cutoff for ingestion, set when shutdown
// begins. It is measured from that moment, not from when an upload starts: an
// upload that gets going late in the grace period must not be given a fresh
// budget, or it would still be running when the server force-closes its
// connection, with no time left to close its crawler run or report progress.
type drainBudget struct {
	cutoff context.Context // cancelled `budget` after shutdown begins
}

// newDrainBudget starts the cutoff timer when shutdown is cancelled. Until
// then it costs nothing: no timer or goroutine runs.
func newDrainBudget(shutdown context.Context, budget time.Duration) *drainBudget {
	cutoff, cancelCutoff := context.WithCancel(context.Background())
	context.AfterFunc(shutdown, func() {
		time.AfterFunc(budget, cancelCutoff)
	})
	return &drainBudget{cutoff: cutoff}
}

// withCutoff returns a context derived from ctx that is also cancelled at the
// cutoff. An upload that starts after the cutoff has passed gets a context
// that is already cancelled. The caller must call the returned cancel func
// when its work ends.
func (b *drainBudget) withCutoff(ctx context.Context) (context.Context, context.CancelFunc) {
	workCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(b.cutoff, cancel)
	return workCtx, func() {
		stop()
		cancel()
	}
}
