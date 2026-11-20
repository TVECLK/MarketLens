package mcpserver

import (
	"context"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// TestDrainBudget runs inside a synctest bubble, where time is fake and only
// advances once every goroutine is blocked. That lets it check the cutoff
// exactly: still live one nanosecond before the budget expires, cancelled at
// expiry. A real-time version could only check a rough window, and would fail
// spuriously whenever the test goroutine woke late.
func TestDrainBudget(t *testing.T) {
	tests := []struct {
		name          string
		budget        time.Duration
		beginShutdown bool
		startDelay    time.Duration // the upload reaches withCutoff this long after shutdown began
		requestEnds   bool          // check that the work context follows the request ending instead
	}{
		{
			name:   "no shutdown: the budget never applies",
			budget: 50 * time.Millisecond,
		},
		{
			name:          "shutdown begins: cut off exactly when the budget expires",
			budget:        300 * time.Millisecond,
			beginShutdown: true,
		},
		{
			name:        "request ends first: work context follows at once",
			budget:      time.Hour,
			requestEnds: true,
		},
		{
			// The budget runs from when shutdown began, not from when this upload
			// started: measured from the upload, it would end at 1.7s.
			name:          "upload starts late: it gets only the time that remains",
			budget:        time.Second,
			beginShutdown: true,
			startDelay:    700 * time.Millisecond,
		},
		{
			name:          "upload starts after the cutoff: cancelled at once",
			budget:        100 * time.Millisecond,
			beginShutdown: true,
			startDelay:    300 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
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
				// What the caller sees the instant withCutoff returns, before any
				// other goroutine has run. The upload handler starts work at once,
				// so a context returned live and cancelled a moment later is not
				// good enough for a late start.
				errAtReturn := work.Err()
				// Record when the work context ended, to make a failure say how
				// early it was cancelled.
				var cancelledAfter atomic.Int64 // nanoseconds since shutdown began; -1 until cancelled
				cancelledAfter.Store(-1)
				context.AfterFunc(work, func() { cancelledAfter.Store(int64(time.Since(shutdownBegan))) })
				synctest.Wait()

				switch {
				case tt.requestEnds:
					if work.Err() != nil {
						t.Fatal("work context was cancelled before the request ended")
					}
					endRequest()
					synctest.Wait()
					if work.Err() == nil {
						t.Error("work context did not follow the request ending")
					}

				case !tt.beginShutdown:
					time.Sleep(24 * time.Hour)
					synctest.Wait()
					if work.Err() != nil {
						t.Error("work context was cancelled although shutdown never began")
					}

				default:
					expiry := shutdownBegan.Add(tt.budget)
					if remaining := time.Until(expiry); remaining <= 0 {
						// The cutoff had already passed: cancelled before withCutoff
						// returned, not on a goroutine that runs afterwards.
						if errAtReturn == nil {
							t.Fatal("work context was still live when withCutoff returned, although the cutoff had passed")
						}
					} else {
						// The cutoff is still ahead: live until the last instant.
						if errAtReturn != nil {
							t.Fatal("work context was cancelled when it started, before the budget expired")
						}
						time.Sleep(remaining - time.Nanosecond)
						synctest.Wait()
						if work.Err() != nil {
							t.Fatalf("work context was cancelled %v after shutdown began, before the %v budget expired",
								time.Duration(cancelledAfter.Load()), tt.budget)
						}
						time.Sleep(time.Nanosecond)
						synctest.Wait()
					}
					// Either the budget has just expired, or it had already passed
					// when the upload started: in both cases it must be cancelled.
					if work.Err() == nil {
						t.Error("work context was not cancelled once the budget had expired")
					}
				}
			})
		})
	}
}
