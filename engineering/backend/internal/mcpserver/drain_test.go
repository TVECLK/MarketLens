package mcpserver

import (
	"context"
	"testing"
	"time"
)

func TestDrainBudget(t *testing.T) {
	tests := []struct {
		name          string
		budget        time.Duration
		beginShutdown bool
		startDelay    time.Duration // the upload reaches withCutoff this long after shutdown began
		requestEnds   bool
		wantCutoff    bool          // the work context must be cancelled within waitFor
		waitFor       time.Duration // how long to wait for the work context to end
		minAfter      time.Duration // earliest allowed cancellation, from when shutdown began
		maxAfter      time.Duration // latest allowed cancellation, from when shutdown began; 0 means no limit
	}{
		{
			name:    "no shutdown: the budget never applies",
			budget:  50 * time.Millisecond,
			waitFor: 200 * time.Millisecond,
		},
		{
			name:          "shutdown begins: cut off once the budget has passed",
			budget:        300 * time.Millisecond,
			beginShutdown: true,
			wantCutoff:    true,
			waitFor:       5 * time.Second,
			minAfter:      250 * time.Millisecond,
			maxAfter:      800 * time.Millisecond,
		},
		{
			name:        "request ends first: work context follows at once",
			budget:      time.Hour,
			requestEnds: true,
			wantCutoff:  true,
			waitFor:     2 * time.Second,
			maxAfter:    500 * time.Millisecond,
		},
		{
			// The budget runs from when shutdown began, not from when this upload
			// started: with the budget measured from the upload, it would end at
			// about 1.7s.
			name:          "upload starts late: it gets only the time that remains",
			budget:        time.Second,
			beginShutdown: true,
			startDelay:    700 * time.Millisecond,
			wantCutoff:    true,
			waitFor:       5 * time.Second,
			minAfter:      900 * time.Millisecond,
			maxAfter:      1350 * time.Millisecond,
		},
		{
			name:          "upload starts after the cutoff: cancelled at once",
			budget:        100 * time.Millisecond,
			beginShutdown: true,
			startDelay:    300 * time.Millisecond,
			wantCutoff:    true,
			waitFor:       2 * time.Second,
			maxAfter:      500 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, endRequest := context.WithCancel(context.Background())
			defer endRequest()
			shutdown, beginShutdown := context.WithCancel(context.Background())
			defer beginShutdown()

			budget := newDrainBudget(shutdown, tt.budget)

			shutdownBegan := time.Now()
			if tt.beginShutdown {
				beginShutdown()
			}
			time.Sleep(tt.startDelay)

			work, cancelWork := budget.withCutoff(request)
			defer cancelWork()
			if tt.requestEnds {
				endRequest()
			}

			select {
			case <-work.Done():
				after := time.Since(shutdownBegan)
				if !tt.wantCutoff {
					t.Fatal("work context was cancelled, want it to stay live")
				}
				if after < tt.minAfter {
					t.Errorf("cancelled %v after shutdown began, want at least %v", after, tt.minAfter)
				}
				if tt.maxAfter > 0 && after > tt.maxAfter {
					t.Errorf("cancelled %v after shutdown began, want within %v", after, tt.maxAfter)
				}
			case <-time.After(tt.waitFor):
				if tt.wantCutoff {
					t.Errorf("work context was not cancelled within %v", tt.waitFor)
				}
			}
		})
	}
}
