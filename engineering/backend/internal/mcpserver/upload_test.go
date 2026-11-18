package mcpserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"marketlens-go-backend/internal/ingestion"
)

// scriptedProcessor stands in for the ingestion service: results decides what
// each ProcessBatch call returns (call is 1-based).
type scriptedProcessor struct {
	calls   int
	results func(call int, jobs []ingestion.RawJobInput) []ingestion.FailedJob
}

func (p *scriptedProcessor) ProcessBatch(_ context.Context, jobs []ingestion.RawJobInput) []ingestion.FailedJob {
	p.calls++
	return p.results(p.calls, jobs)
}

func TestProcessWithRetry(t *testing.T) {
	jobs := []ingestion.RawJobInput{{JobID: "j1"}, {JobID: "j2"}}
	transient := func(id string) ingestion.FailedJob {
		return ingestion.FailedJob{JobID: id, StatusCode: http.StatusServiceUnavailable, Error: "llm busy"}
	}
	permanent := func(id string) ingestion.FailedJob {
		return ingestion.FailedJob{JobID: id, StatusCode: http.StatusUnprocessableEntity, Error: "unextractable"}
	}
	quick := func(int) time.Duration { return time.Millisecond }

	tests := []struct {
		name            string
		results         func(call int, jobs []ingestion.RawJobInput) []ingestion.FailedJob
		delay           func(int) time.Duration
		shutdownBegun   bool
		cancelCtxAfter  time.Duration // cancel the work context this long after starting; 0 means never
		wantCalls       int
		wantFailures    int
		wantInterrupted bool
	}{
		{
			name:         "all jobs succeed first time",
			results:      func(int, []ingestion.RawJobInput) []ingestion.FailedJob { return nil },
			delay:        quick,
			wantCalls:    1,
			wantFailures: 0,
		},
		{
			name: "transient failure succeeds on retry",
			results: func(call int, _ []ingestion.RawJobInput) []ingestion.FailedJob {
				if call == 1 {
					return []ingestion.FailedJob{transient("j1")}
				}
				return nil
			},
			delay:        quick,
			wantCalls:    2,
			wantFailures: 0,
		},
		{
			name: "permanent failure is never retried",
			results: func(int, []ingestion.RawJobInput) []ingestion.FailedJob {
				return []ingestion.FailedJob{permanent("j1")}
			},
			delay:        quick,
			wantCalls:    1,
			wantFailures: 1,
		},
		{
			name: "transient failure on every attempt gives up after the retry limit",
			results: func(int, []ingestion.RawJobInput) []ingestion.FailedJob {
				return []ingestion.FailedJob{transient("j1")}
			},
			delay:        quick,
			wantCalls:    1 + manualUploadMaxRetries,
			wantFailures: 1,
		},
		{
			name: "permanent failure is kept while the transient one is retried",
			results: func(call int, _ []ingestion.RawJobInput) []ingestion.FailedJob {
				if call == 1 {
					return []ingestion.FailedJob{permanent("j1"), transient("j2")}
				}
				return nil
			},
			delay:        quick,
			wantCalls:    2,
			wantFailures: 1,
		},
		{
			name: "shutdown begun: transient failure is not retried",
			results: func(int, []ingestion.RawJobInput) []ingestion.FailedJob {
				return []ingestion.FailedJob{transient("j1")}
			},
			delay:           quick,
			shutdownBegun:   true,
			wantCalls:       1,
			wantFailures:    1,
			wantInterrupted: true,
		},
		{
			name:          "shutdown begun but nothing needs a retry: not interrupted",
			results:       func(int, []ingestion.RawJobInput) []ingestion.FailedJob { return nil },
			delay:         quick,
			shutdownBegun: true,
			wantCalls:     1,
			wantFailures:  0,
		},
		{
			name: "context ends during the retry wait",
			results: func(int, []ingestion.RawJobInput) []ingestion.FailedJob {
				return []ingestion.FailedJob{transient("j1")}
			},
			delay:           func(int) time.Duration { return time.Hour },
			cancelCtxAfter:  20 * time.Millisecond,
			wantCalls:       1,
			wantFailures:    1,
			wantInterrupted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancelCtx := context.WithCancel(context.Background())
			defer cancelCtx()
			if tt.cancelCtxAfter > 0 {
				time.AfterFunc(tt.cancelCtxAfter, cancelCtx)
			}
			shutdown, beginShutdown := context.WithCancel(context.Background())
			defer beginShutdown()
			if tt.shutdownBegun {
				beginShutdown()
			}

			proc := &scriptedProcessor{results: tt.results}
			failures, interrupted := processWithRetry(ctx, shutdown, proc, jobs, tt.delay)

			if proc.calls != tt.wantCalls {
				t.Errorf("ProcessBatch calls = %d, want %d", proc.calls, tt.wantCalls)
			}
			if len(failures) != tt.wantFailures {
				t.Errorf("failures = %d (%v), want %d", len(failures), failures, tt.wantFailures)
			}
			if interrupted != tt.wantInterrupted {
				t.Errorf("interrupted = %v, want %v", interrupted, tt.wantInterrupted)
			}
		})
	}
}

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

func TestNewSubmitOutput(t *testing.T) {
	jobsByID := map[string]newspaperJob{
		"a": {Employer: "Acme", JobRole: "Engineer"},
		"b": {Employer: "Beta", JobRole: "Clerk"},
		"c": {Employer: "Gamma", JobRole: "Driver"},
	}
	oneFailure := []ingestion.FailedJob{{JobID: "b", Error: "unextractable"}}

	tests := []struct {
		name            string
		failures        []ingestion.FailedJob
		interrupted     bool
		wantSaved       int
		wantFailedCount int
		wantMessage     []string
		wantFailedRoles []string
	}{
		{
			name:        "everything saved",
			wantSaved:   3,
			wantMessage: []string{"Vacancies processed"},
		},
		{
			name:            "one job failed",
			failures:        oneFailure,
			wantSaved:       2,
			wantFailedCount: 1,
			wantMessage:     []string{"Vacancies processed", "were not saved"},
			wantFailedRoles: []string{"Clerk"},
		},
		{
			name:            "interrupted upload tells the client to resubmit everything",
			failures:        oneFailure,
			interrupted:     true,
			wantSaved:       2,
			wantFailedCount: 1,
			wantMessage:     []string{"stopped before it finished", "Submit the full list", "Do not submit only the failed jobs"},
			wantFailedRoles: []string{"Clerk"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := newSubmitOutput("Daily News", 3, tt.failures, jobsByID, tt.interrupted)

			if out.Source != "Daily News" || out.Submitted != 3 {
				t.Errorf("source/submitted = %q/%d, want Daily News/3", out.Source, out.Submitted)
			}
			if out.SavedCount != tt.wantSaved || out.FailedCount != tt.wantFailedCount {
				t.Errorf("saved/failed = %d/%d, want %d/%d", out.SavedCount, out.FailedCount, tt.wantSaved, tt.wantFailedCount)
			}
			if out.Interrupted != tt.interrupted {
				t.Errorf("interrupted = %v, want %v", out.Interrupted, tt.interrupted)
			}
			for _, want := range tt.wantMessage {
				if !strings.Contains(out.Message, want) {
					t.Errorf("message %q does not contain %q", out.Message, want)
				}
			}
			var roles []string
			for _, f := range out.Failed {
				roles = append(roles, f.JobRole)
			}
			if strings.Join(roles, ",") != strings.Join(tt.wantFailedRoles, ",") {
				t.Errorf("failed roles = %v, want %v", roles, tt.wantFailedRoles)
			}
		})
	}
}
