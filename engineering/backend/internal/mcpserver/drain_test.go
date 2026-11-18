package mcpserver

import (
	"context"
	"testing"
	"time"
)

func TestWithDrainBudget(t *testing.T) {
	tests := []struct {
		name            string
		budget          time.Duration
		shutdownBegins  bool
		requestEnds     bool
		wantLiveInitial bool          // must still be live shortly after the trigger
		wantCancelled   bool          // must be cancelled within waitFor
		waitFor         time.Duration // how long to wait for the work context to end
	}{
		{
			name:          "no shutdown: the budget never applies",
			budget:        50 * time.Millisecond,
			wantCancelled: false,
			waitFor:       200 * time.Millisecond,
		},
		{
			name:            "shutdown begins: cut off only after the budget",
			budget:          300 * time.Millisecond,
			shutdownBegins:  true,
			wantLiveInitial: true,
			wantCancelled:   true,
			waitFor:         5 * time.Second,
		},
		{
			name:          "request ends first: work context follows at once",
			budget:        time.Hour,
			requestEnds:   true,
			wantCancelled: true,
			waitFor:       2 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, endRequest := context.WithCancel(context.Background())
			defer endRequest()
			shutdown, beginShutdown := context.WithCancel(context.Background())
			defer beginShutdown()

			work, cancelWork := withDrainBudget(request, shutdown, tt.budget)
			defer cancelWork()

			if tt.shutdownBegins {
				beginShutdown()
			}
			if tt.requestEnds {
				endRequest()
			}

			if tt.wantLiveInitial {
				time.Sleep(30 * time.Millisecond)
				if work.Err() != nil {
					t.Fatal("work context was cancelled before the budget ran out")
				}
			}

			select {
			case <-work.Done():
				if !tt.wantCancelled {
					t.Error("work context was cancelled, want it to stay live")
				}
			case <-time.After(tt.waitFor):
				if tt.wantCancelled {
					t.Errorf("work context was not cancelled within %v", tt.waitFor)
				}
			}
		})
	}
}
