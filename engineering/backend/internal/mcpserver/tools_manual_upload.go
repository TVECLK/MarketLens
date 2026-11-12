package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/service"
)

const submitNewspaperVacanciesScope = "submit-newspaper-vacancies"

type newspaperJob struct {
	Employer    string `json:"employer" jsonschema:"employer/company name as printed in the ad"`
	JobRole     string `json:"job_role" jsonschema:"job title / role as printed in the ad"`
	Location    string `json:"location" jsonschema:"work location, city or region"`
	Description string `json:"description" jsonschema:"job description: responsibilities, requirements, benefits"`
}

type submitNewspaperVacanciesInput struct {
	Source        string         `json:"source" jsonschema:"name of the manual source the user picked via select_manual_source, exactly as returned (e.g. 'Daily News')"`
	Jobs          []newspaperJob `json:"jobs" jsonschema:"list of job vacancies extracted from the uploaded newspaper image"`
	UserConfirmed bool           `json:"user_confirmed,omitempty" jsonschema:"only needed when the client cannot show the confirmation form (the tool then tells you so): set to true ONLY after you have shown the jobs and source to the user and they explicitly approved them. Never set this on your own."`
}

// submitDecisionOutput is returned instead of a save result when the user
// answers the confirmation form with No or Edit.
type submitDecisionOutput struct {
	Status  string `json:"status" jsonschema:"cancelled or edits_requested"`
	Edits   string `json:"edits,omitempty" jsonschema:"the changes the user asked for, in their own words"`
	Message string `json:"message"`
}

const (
	confirmYes  = "Yes"
	confirmNo   = "No"
	confirmEdit = "Edit"
)

// clientSupportsElicitation reports whether the connected client can show a
// form to the user. It follows the SDK's own rule in ServerSession.Elicit:
// a client declaring only URL elicitation cannot do forms, while an empty
// elicitation declaration ({}) is legacy and means form support.
func clientSupportsElicitation(req *mcp.CallToolRequest) bool {
	if req == nil || req.Session == nil {
		return false
	}
	ip := req.Session.InitializeParams()
	if ip == nil || ip.Capabilities == nil || ip.Capabilities.Elicitation == nil {
		return false
	}
	caps := ip.Capabilities.Elicitation
	return caps.Form != nil || caps.URL == nil
}

// confirmationMessage renders the jobs and source for the confirmation form.
func confirmationMessage(source string, jobs []newspaperJob) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Save these %d vacancies from %q?\n\n", len(jobs), source)
	for i, j := range jobs {
		desc := []rune(strings.TrimSpace(j.Description))
		if len(desc) > 200 {
			desc = append(desc[:200], '…')
		}
		fmt.Fprintf(&b, "%d. %s - %s (%s)\n   %s\n", i+1, j.JobRole, j.Employer, j.Location, string(desc))
	}
	return b.String()
}

// confirmSubmission shows a Yes / No / Edit select form. It returns the
// chosen option (confirmYes, confirmNo or confirmEdit) and, for Edit, the
// edits the user typed in a follow-up form. A declined or dismissed form
// counts as No, so nothing is ever saved without an explicit Yes.
func confirmSubmission(ctx context.Context, session *mcp.ServerSession, source string, jobs []newspaperJob) (choice, edits string, err error) {
	res, err := session.Elicit(ctx, &mcp.ElicitParams{
		Message: confirmationMessage(source, jobs),
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"decision": map[string]any{
					"type":  "string",
					"title": "Save these vacancies?",
					"enum":  []string{confirmYes, confirmNo, confirmEdit},
				},
			},
			"required": []string{"decision"},
		},
	})
	if err != nil {
		return "", "", err
	}
	if res.Action != "accept" {
		return confirmNo, "", nil
	}
	picked, _ := res.Content["decision"].(string)
	switch {
	case strings.EqualFold(picked, confirmYes):
		return confirmYes, "", nil
	case strings.EqualFold(picked, confirmEdit):
		// fall through to ask what to change
	default:
		return confirmNo, "", nil
	}

	res, err = session.Elicit(ctx, &mcp.ElicitParams{
		Message: "What edits do you want to make to the extracted vacancies?",
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"edits": map[string]any{
					"type":        "string",
					"title":       "Edits",
					"description": "Describe the changes, e.g. 'job 2 location is Kandy; remove job 4'",
				},
			},
			"required": []string{"edits"},
		},
	})
	if err != nil {
		return "", "", err
	}
	if res.Action == "accept" {
		edits, _ = res.Content["edits"].(string)
	}
	return confirmEdit, strings.TrimSpace(edits), nil
}

type manualSourceOutput struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

const noMatchingSourceMessage = "If there is no matching source, you should add source from the dashboard and then try again."

type selectManualSourceOutput struct {
	Status  string               `json:"status" jsonschema:"selected, choose_in_chat, cancelled or no_sources"`
	Source  string               `json:"source,omitempty"`
	Sources []manualSourceOutput `json:"sources,omitempty"`
	Message string               `json:"message"`
}

type failedVacancyOutput struct {
	JobRole  string `json:"job_role"`
	Employer string `json:"employer"`
	Error    string `json:"error"`
}

type submitNewspaperVacanciesOutput struct {
	Message     string                `json:"message"`
	Source      string                `json:"source"`
	Submitted   int                   `json:"submitted"`
	FailedCount int                   `json:"failed_count"`
	Failed      []failedVacancyOutput `json:"failed,omitempty"`

	ReconciledStaleJobs int64  `json:"reconciled_stale_jobs"`
	ReconcileNote       string `json:"reconcile_note,omitempty"`
}

// hasScope reports whether the verified token on ctx carries scope.
func hasScope(ctx context.Context, scope string) bool {
	scopes, _ := ctx.Value(scopesKey).([]string)
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

const (
	// manualUploadMaxRetries is how many times jobs that failed transiently
	// are re-submitted after the first attempt before they are given up on.
	manualUploadMaxRetries = 3
	manualUploadRetryDelay = time.Second
)

// processWithRetry runs the batch, then re-runs only the jobs that failed
// with a transient (503) error, up to manualUploadMaxRetries times with a
// growing delay. Permanent failures (4xx: bad input, unextractable job) are
// never retried. Whatever still fails afterwards is returned for reporting.
func processWithRetry(ctx context.Context, svc *ingestion.IngestionService, jobs []ingestion.RawJobInput) []ingestion.FailedJob {
	failures := svc.ProcessBatch(ctx, jobs)
	byID := make(map[string]ingestion.RawJobInput, len(jobs))
	for _, j := range jobs {
		byID[j.JobID] = j
	}

	for attempt := 1; attempt <= manualUploadMaxRetries; attempt++ {
		var retry []ingestion.RawJobInput
		var permanent []ingestion.FailedJob
		for _, f := range failures {
			if f.StatusCode == http.StatusServiceUnavailable {
				retry = append(retry, byID[f.JobID])
			} else {
				permanent = append(permanent, f)
			}
		}
		if len(retry) == 0 {
			return failures
		}

		select {
		case <-ctx.Done():
			return failures
		case <-time.After(time.Duration(attempt) * manualUploadRetryDelay):
		}

		log.Printf("retrying %d transiently failed manual job(s), attempt %d/%d", len(retry), attempt, manualUploadMaxRetries)
		failures = append(permanent, svc.ProcessBatch(ctx, retry)...)
	}
	return failures
}

func registerManualUploadTools(server *mcp.Server, svcs *service.Services, ingestionSvc *ingestion.IngestionService) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "select_manual_source",
		Description: "Let the user choose which manual source (newspaper) an uploaded newspaper image came from. " +
			"Call this after extracting the jobs from the image. It shows the user a select box of the active " +
			"manual sources; pass the returned 'source' to submit_newspaper_vacancies. If the client cannot show " +
			"a select box, the result has status 'choose_in_chat' and a list of sources: ask the user to pick one " +
			"of exactly those names. The user may only choose a listed source - never invent or guess one. " +
			"If no source matches, relay the 'message' to the user.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		active := true
		sources, err := svcs.Sources.ListSources(ctx, string(models.SourceTypeManual), &active)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list manual sources: %w", err)
		}
		if len(sources) == 0 {
			return nil, selectManualSourceOutput{
				Status:  "no_sources",
				Message: "There are no active manual sources. Add a source from the dashboard and then try again.",
			}, nil
		}

		names := make([]string, len(sources))
		list := make([]manualSourceOutput, len(sources))
		for i, s := range sources {
			names[i] = s.Name
			list[i] = manualSourceOutput{ID: s.ID, Name: s.Name}
		}

		// Show a real select box when the client supports elicitation.
		if clientSupportsElicitation(req) {
			res, err := req.Session.Elicit(ctx, &mcp.ElicitParams{
				Message: "Select the newspaper this image came from. " + noMatchingSourceMessage,
				RequestedSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"source": map[string]any{
							"type":  "string",
							"title": "Source",
							"enum":  names,
						},
					},
					"required": []string{"source"},
				},
			})
			if err == nil {
				if res.Action != "accept" {
					return nil, selectManualSourceOutput{
						Status:  "cancelled",
						Message: "The user did not select a source. " + noMatchingSourceMessage,
					}, nil
				}
				if chosen, _ := res.Content["source"].(string); slices.Contains(names, chosen) {
					return nil, selectManualSourceOutput{Status: "selected", Source: chosen, Message: "Source selected."}, nil
				}
			}
			// Elicitation unsupported or failed: fall through to the chat list.
		}

		return nil, selectManualSourceOutput{
			Status:  "choose_in_chat",
			Sources: list,
			Message: "Ask the user to pick one of these sources. " + noMatchingSourceMessage,
		}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "submit_newspaper_vacancies",
		Description: "Submit job vacancies extracted from a newspaper image for deduplication, " +
			"classification, and storage in the labour market database. This tool does not read " +
			"images itself - before calling it, read the uploaded newspaper image yourself and " +
			"extract each job's employer, role, location, and description.\n\n" +
			"Workflow: (1) extract the jobs from the image, (2) call select_manual_source so the " +
			"user chooses the newspaper as the source, (3) show the extracted jobs and the chosen " +
			"source in the chat as a readable list or table, (4) call this tool. It shows the user a " +
			"Yes / No / Edit form and saves only on Yes.\n\n" +
			"Results: status 'cancelled' means the user chose No - stop, nothing was saved. Status " +
			"'edits_requested' means the user chose Edit - apply their 'edits' to the jobs (if 'edits' " +
			"is empty, ask the user in chat what to change), show the updated jobs in the chat, then " +
			"call this tool again with the updated jobs so they confirm again.\n\n" +
			"If the client cannot show the form, this tool says so: then ask the user in chat and call " +
			"it again with user_confirmed=true only after their explicit yes.\n\n" +
			"IMPORTANT: never invent a source name - it must be one returned by select_manual_source.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in submitNewspaperVacanciesInput) (*mcp.CallToolResult, any, error) {
		// Checked first so the user is never asked to confirm a submission
		// that cannot be processed.
		if ingestionSvc == nil {
			return nil, nil, fmt.Errorf("newspaper vacancy submission is unavailable: the LLM is not configured on the MCP server.")
		}

		if len(in.Jobs) == 0 {
			return nil, nil, fmt.Errorf("jobs list must not be empty - extract at least one job from the image first")
		}

		if !hasScope(ctx, submitNewspaperVacanciesScope) {
			return nil, nil, fmt.Errorf("missing required scope: %s", submitNewspaperVacanciesScope)
		}

		sourceName := strings.TrimSpace(in.Source)
		if sourceName == "" {
			return nil, nil, fmt.Errorf("source is required - call select_manual_source and use the source it returns")
		}
		source, err := svcs.Sources.GetSourceByName(ctx, sourceName)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("source %q is not registered - call select_manual_source and use the source it returns. %s", sourceName, noMatchingSourceMessage)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to look up source %q: %w", sourceName, err)
		}
		if source.Type != models.SourceTypeManual {
			return nil, nil, fmt.Errorf("source %q is not a manual source - call select_manual_source and use the source it returns. %s", sourceName, noMatchingSourceMessage)
		}
		if !source.IsActive {
			return nil, nil, fmt.Errorf("source %q is disabled", sourceName)
		}

		// Confirm only after the cheap checks above, so the user is never asked
		// to approve a submission that would be rejected anyway.
		if clientSupportsElicitation(req) {
			choice, edits, err := confirmSubmission(ctx, req.Session, source.Name, in.Jobs)
			if err != nil {
				return nil, nil, fmt.Errorf("could not get the user's confirmation, nothing was saved: %w", err)
			}
			switch choice {
			case confirmNo:
				return nil, submitDecisionOutput{
					Status:  "cancelled",
					Message: "The user chose No. The submission was cancelled and nothing was saved.",
				}, nil
			case confirmEdit:
				msg := "The user chose Edit. Apply their edits to the jobs, show the updated jobs in the chat, then call this tool again with the updated jobs."
				if edits == "" {
					msg = "The user chose Edit but gave no details. Ask them in the chat what to change, update the jobs, show them in the chat, then call this tool again."
				}
				return nil, submitDecisionOutput{Status: "edits_requested", Edits: edits, Message: msg}, nil
			}
		} else if !in.UserConfirmed {
			return nil, nil, fmt.Errorf(
				"this client cannot show a confirmation form: show the extracted jobs to the user in the chat, " +
					"get their explicit yes, then call this tool again with user_confirmed=true")
		}

		// The ingestion pipeline links every job to a crawler run, so each
		// manual submission gets its own run, closed once the batch is done.
		run, err := svcs.CrawlerRuns.CreateCrawlerRun(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to start ingestion run: %w", err)
		}
		runStatus := "FAILED"
		defer func() {
			// Detached from ctx so a cancelled request still closes the run.
			if err := svcs.CrawlerRuns.CompleteCrawlerRun(context.WithoutCancel(ctx), run.ID, runStatus); err != nil {
				log.Printf("failed to close crawler run %d as %s: %v", run.ID, runStatus, err)
			}
		}()

		rawJobs := make([]ingestion.RawJobInput, len(in.Jobs))
		jobsByID := make(map[string]newspaperJob, len(in.Jobs))
		for i, j := range in.Jobs {
			id := fmt.Sprintf("manual-%d-%d", run.ID, i)
			rawJobs[i] = ingestion.RawJobInput{
				JobID:        id,
				Employer:     j.Employer,
				JobRole:      j.JobRole,
				Location:     j.Location,
				Description:  j.Description,
				CrawlerRunID: run.ID,
				Source:       source.Name,
			}
			jobsByID[id] = j
		}

		failures := processWithRetry(ctx, ingestionSvc, rawJobs)
		runStatus = "COMPLETED"

		out := submitNewspaperVacanciesOutput{
			Message:     "Vacancies processed. Jobs not listed under 'failed' were saved or matched an existing duplicate.",
			Source:      source.Name,
			Submitted:   len(rawJobs),
			FailedCount: len(failures),
		}

		closed, err := svcs.JobPosts.ReconcileStaleVacancies(ctx, run.ID, []string{source.Name})
		if err != nil {
			log.Printf("reconcile after manual run %d failed: %v", run.ID, err)
			out.ReconcileNote = "vacancies were saved, but closing stale vacancies failed"
		}
		out.ReconciledStaleJobs = closed
		for _, f := range failures {
			j := jobsByID[f.JobID]
			out.Failed = append(out.Failed, failedVacancyOutput{JobRole: j.JobRole, Employer: j.Employer, Error: f.Error})
		}
		return nil, out, nil
	})
}
