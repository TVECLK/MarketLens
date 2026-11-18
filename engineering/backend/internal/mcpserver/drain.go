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

// withDrainBudget returns a context derived from ctx that is also cancelled
// budget after shutdown is cancelled. Until shutdown begins it behaves like
// ctx. The caller must call the returned cancel func when its work ends.
func withDrainBudget(ctx, shutdown context.Context, budget time.Duration) (context.Context, context.CancelFunc) {
	workCtx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-shutdown.Done():
			select {
			case <-time.After(budget):
				cancel()
			case <-workCtx.Done():
			}
		case <-workCtx.Done():
		}
	}()
	return workCtx, cancel
}
