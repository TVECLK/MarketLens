package mcpserver

import (
	"context"
	"marketlens-go-backend/internal/ingestion"
	"net/http"
	"strings"
	"testing"
	"time"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHasScope(t *testing.T) {
	withScopes := func(scopes ...string) *mcp.CallToolRequest {
		return &mcp.CallToolRequest{Extra: &mcp.RequestExtra{TokenInfo: &mcpauth.TokenInfo{Scopes: scopes}}}
	}
	tests := []struct {
		name string
		req  *mcp.CallToolRequest
		want bool
	}{
		{"nil request", nil, false},
		{"no extra", &mcp.CallToolRequest{}, false},
		{"no token info", &mcp.CallToolRequest{Extra: &mcp.RequestExtra{}}, false},
		{"scope present", withScopes("openid", submitNewspaperVacanciesScope), true},
		{"scope absent", withScopes("openid"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasScope(tt.req, submitNewspaperVacanciesScope); got != tt.want {
				t.Errorf("hasScope = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientSupportsElicitation(t *testing.T) {
	handler := func(context.Context, *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
		return &mcp.ElicitResult{Action: "decline"}, nil
	}
	tests := []struct {
		name string
		opts *mcp.ClientOptions
		want bool
	}{
		{"no elicitation capability", nil, false},
		{"elicitation handler set", &mcp.ClientOptions{ElicitationHandler: handler}, true},
		{"form elicitation declared", &mcp.ClientOptions{
			ElicitationHandler: handler,
			Capabilities:       &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{Form: &mcp.FormElicitationCapabilities{}}},
		}, true},
		{"only URL elicitation declared", &mcp.ClientOptions{
			ElicitationHandler: handler,
			Capabilities:       &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{URL: &mcp.URLElicitationCapabilities{}}},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
			mcp.AddTool(server, &mcp.Tool{Name: "probe"},
				func(_ context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
					return nil, map[string]any{"supports": clientSupportsElicitation(req)}, nil
				})
			session := connectInMemory(t, server, tt.opts)

			res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "probe"})
			if err != nil {
				t.Fatalf("call tool: %v", err)
			}
			if got := boolField(t, res, "supports"); got != tt.want {
				t.Errorf("clientSupportsElicitation = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("no request or session", func(t *testing.T) {
		if clientSupportsElicitation(nil) {
			t.Error("nil request reported as supporting elicitation")
		}
		if clientSupportsElicitation(&mcp.CallToolRequest{}) {
			t.Error("request without a session reported as supporting elicitation")
		}
	})
}

func TestConfirmationMessage(t *testing.T) {
	longDescription := strings.Repeat("Responsible for planning and delivery. ", 12) + "END-OF-DESCRIPTION"

	tests := []struct {
		name       string
		source     string
		jobs       []newspaperJob
		want       []string
		wantAbsent []string
	}{
		{
			name:       "long description is shown in full",
			source:     "Daily News",
			jobs:       []newspaperJob{{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: longDescription}},
			want:       []string{`Save these 1 vacancies from "Daily News"?`, "1. Engineer - Acme (Colombo)", "END-OF-DESCRIPTION"},
			wantAbsent: []string{"…"},
		},
		{
			name:   "jobs are numbered in order",
			source: "Sunday Times",
			jobs: []newspaperJob{
				{Employer: "A", JobRole: "Clerk", Location: "Kandy", Description: "first"},
				{Employer: "B", JobRole: "Driver", Location: "Galle", Description: "second"},
			},
			want: []string{`Save these 2 vacancies from "Sunday Times"?`, "1. Clerk - A (Kandy)", "2. Driver - B (Galle)", "first", "second"},
		},
		{
			name:   "description whitespace is trimmed",
			source: "Daily News",
			jobs:   []newspaperJob{{Employer: "A", JobRole: "Clerk", Location: "Kandy", Description: "  \n padded text \n "}},
			want:   []string{"\n   padded text\n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := confirmationMessage(tt.source, tt.jobs)
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("message does not contain %q:\n%s", w, got)
				}
			}
			for _, w := range tt.wantAbsent {
				if strings.Contains(got, w) {
					t.Errorf("message contains %q:\n%s", w, got)
				}
			}
		})
	}
}

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		attempt int
		base    time.Duration // delay before jitter
	}{
		{1, manualUploadRetryBaseDelay},
		{2, 2 * manualUploadRetryBaseDelay},
		{3, 4 * manualUploadRetryBaseDelay},
	}
	for _, tt := range tests {
		t.Run("attempt "+string(rune('0'+tt.attempt)), func(t *testing.T) {
			low := time.Duration(float64(tt.base) * (1 - manualUploadRetryJitter))
			high := time.Duration(float64(tt.base) * (1 + manualUploadRetryJitter))
			for range 500 {
				if got := retryDelay(tt.attempt); got < low || got > high {
					t.Fatalf("retryDelay(%d) = %v, want within [%v, %v]", tt.attempt, got, low, high)
				}
			}
		})
	}
}

func TestConfirmUnlessShuttingDown(t *testing.T) {
	tests := []struct {
		name                  string
		userAnswersYes        bool // the client answers the form at once
		shutdownBefore        bool // shutdown has already begun when the tool starts waiting
		shutdownWhenFormShown bool // shutdown begins once the client has received the form, while it is unanswered
		wantShuttingDown      bool
		wantConfirmed         bool
	}{
		{name: "user answers Yes before any shutdown", userAnswersYes: true, wantConfirmed: true},
		{name: "shutdown begins while the form is open and unanswered", shutdownWhenFormShown: true, wantShuttingDown: true},
		{name: "shutdown already begun when the tool starts waiting", shutdownBefore: true, wantShuttingDown: true},
	}
	jobs := []newspaperJob{{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Builds things"}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdown, beginShutdown := context.WithCancel(context.Background())
			defer beginShutdown()

			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
			mcp.AddTool(server, &mcp.Tool{Name: "confirm_probe"},
				func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
					choice, _, shuttingDown, err := confirmUnlessShuttingDown(ctx, shutdown, req.Session, "Daily News", jobs)
					if err != nil {
						return nil, nil, err
					}
					return nil, map[string]any{"shutting_down": shuttingDown, "confirmed": choice == confirmYes}, nil
				})

			// A client that never answers keeps the form open until the test ends.
			// formShown is signalled when the form reaches the client, so shutdown
			// can be started at that moment and not after a guessed delay.
			release := make(chan struct{})
			defer close(release)
			formShown := make(chan struct{}, 1)
			session := connectInMemory(t, server, &mcp.ClientOptions{
				ElicitationHandler: func(ctx context.Context, _ *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
					select {
					case formShown <- struct{}{}:
					default:
					}
					if tt.userAnswersYes {
						return &mcp.ElicitResult{Action: "accept", Content: map[string]any{"decision": confirmYes}}, nil
					}
					select {
					case <-release:
					case <-ctx.Done():
					}
					return &mcp.ElicitResult{Action: "decline"}, nil
				},
			})

			callCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Measured from the moment shutdown begins, or from the call when no
			// shutdown is involved.
			started := time.Now()
			if tt.shutdownBefore {
				beginShutdown()
			}

			type callOutcome struct {
				res *mcp.CallToolResult
				err error
			}
			outcome := make(chan callOutcome, 1)
			go func() {
				res, err := session.CallTool(callCtx, &mcp.CallToolParams{Name: "confirm_probe"})
				outcome <- callOutcome{res: res, err: err}
			}()

			if tt.shutdownWhenFormShown {
				select {
				case <-formShown:
				case <-time.After(3 * time.Second):
					t.Fatal("the form never reached the client, so there is no open form to shut down under")
				}
				started = time.Now()
				beginShutdown()
			}

			var res *mcp.CallToolResult
			select {
			case got := <-outcome:
				if got.err != nil {
					t.Fatalf("call tool: %v (the tool kept waiting for an answer that cannot arrive)", got.err)
				}
				res = got.res
			case <-time.After(4 * time.Second):
				t.Fatal("the tool did not return: it kept waiting for an answer that cannot arrive")
			}
			if elapsed := time.Since(started); elapsed > 2*time.Second {
				t.Errorf("tool took %v to return", elapsed)
			}

			if got := boolField(t, res, "shutting_down"); got != tt.wantShuttingDown {
				t.Errorf("shutting_down = %v, want %v", got, tt.wantShuttingDown)
			}
			if got := boolField(t, res, "confirmed"); got != tt.wantConfirmed {
				t.Errorf("confirmed = %v, want %v", got, tt.wantConfirmed)
			}
		})
	}
}

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

func TestFinalRunStatus(t *testing.T) {
	tests := []struct {
		name        string
		interrupted bool
		savedCount  int
		want        string
	}{
		{"every job saved", false, 3, "COMPLETED"},
		{"every job matched an existing duplicate", false, 3, "COMPLETED"},
		{"some jobs failed, some succeeded", false, 1, "COMPLETED"},
		{"every job failed", false, 0, "FAILED"},
		{"interrupted with some jobs saved", true, 2, "FAILED"},
		{"interrupted with nothing saved", true, 0, "FAILED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := finalRunStatus(tt.interrupted, tt.savedCount); got != tt.want {
				t.Errorf("finalRunStatus(interrupted=%v, saved=%d) = %q, want %q", tt.interrupted, tt.savedCount, got, tt.want)
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
