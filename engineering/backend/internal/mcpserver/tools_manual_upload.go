package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type newspaperJob struct {
	Employer    string `json:"employer" jsonschema:"employer/company name as printed in the ad"`
	JobRole     string `json:"job_role" jsonschema:"job title / role as printed in the ad"`
	Location    string `json:"location" jsonschema:"work location, city or region"`
	Description string `json:"description" jsonschema:"job description: responsibilities, requirements, benefits"`
	Source      string `json:"source" jsonschema:"name of the newspaper this ad was found in, e.g. 'Daily News'"`
}

type submitNewspaperVacanciesInput struct {
	Jobs          []newspaperJob `json:"jobs" jsonschema:"list of job vacancies extracted from the uploaded newspaper image"`
	UserConfirmed bool           `json:"user_confirmed" jsonschema:"set to true ONLY after you have shown the extracted jobs to the user and they explicitly approved them. Never set this on your own."`
}

func registerManualUploadTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "submit_newspaper_vacancies",
		Description: "Submit job vacancies extracted from a newspaper image for deduplication, " +
			"classification, and storage in the labour market database. This tool does not read " +
			"images itself - before calling it, read the uploaded newspaper image yourself and " +
			"extract each job's employer, role, location, description, and the newspaper's name " +
			"as the source.\n\n" +
			"IMPORTANT: Do not call this tool immediately after extraction. First present the " +
			"extracted jobs to the user as a readable list or table, ask them to confirm or " +
			"correct the data, and only call this tool once the user has explicitly approved.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in submitNewspaperVacanciesInput) (*mcp.CallToolResult, any, error) {
		if len(in.Jobs) == 0 {
			return nil, nil, fmt.Errorf("jobs list must not be empty - extract at least one job from the image first")
		}

		if !in.UserConfirmed {
			return nil, nil, fmt.Errorf(
				"user_confirmed is false: show the extracted jobs to the user for review, " +
					"get explicit approval, then call this tool again with user_confirmed=true")
		}

		scopes, _ := ctx.Value(scopesKey).([]string)
		hasScope := false
		for _, s := range scopes {
			if s == "submit-newspaper-vacancies" {
				hasScope = true
				break
			}
		}
		if !hasScope {
			return nil, nil, fmt.Errorf("missing required scope: submit-newspaper-vacancies")
		}

		body, err := json.Marshal(in.Jobs)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode jobs: %w", err)
		}

		crawlerURL := os.Getenv("CRAWLER_API_URL")
		if crawlerURL == "" {
			crawlerURL = "http://crawler:8000"
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, crawlerURL+"/manual-upload-jobs", bytes.NewReader(body))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to reach crawler service: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read crawler response: %w", err)
		}

		var result map[string]any
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, nil, fmt.Errorf("invalid response from crawler: %s", string(respBody))
		}

		if resp.StatusCode >= 400 {
			return nil, result, fmt.Errorf("crawler service returned status %d", resp.StatusCode)
		}
		return nil, result, nil
	})
}