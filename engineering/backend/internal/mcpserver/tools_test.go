package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"
)

// connectInMemory connects a client to server over an in-memory transport.
func connectInMemory(t *testing.T, server *mcp.Server, opts *mcp.ClientOptions) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, opts)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

// boolField returns the boolean field key from a tool result's structured
// content. It fails the test if the content is not an object, the field is
// missing, or it is not a boolean, so a malformed response cannot pass as a
// false value.
func boolField(t *testing.T, res *mcp.CallToolResult, key string) bool {
	t.Helper()
	out, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content is %T, want an object", res.StructuredContent)
	}
	value, present := out[key]
	if !present {
		t.Fatalf("result has no %q field: %v", key, out)
	}
	b, ok := value.(bool)
	if !ok {
		t.Fatalf("%q is %T (%v), want a bool", key, value, value)
	}
	return b
}

// TestNew_RegistersTools builds the real server and lists its tools. Building
// it also runs every AddTool call, which panics on a bad input or output
// schema, so a broken jsonschema tag fails here and not at startup. Registering
// needs no database: handlers only run when a tool is called.
func TestNew_RegistersTools(t *testing.T) {
	wantTools := []string{
		// reference lists
		"get_experiences", "get_provinces", "get_job_types", "get_employment_sectors",
		"get_education_levels", "get_formalities", "get_genders", "get_vocational_educations",
		// occupation hierarchy
		"get_major_groups", "get_major_group", "get_sub_major_groups", "get_sub_major_group",
		"get_minor_groups", "get_minor_group", "get_unit_groups", "get_unit_group",
		"get_occupation_groups", "get_occupation_group",
		// industry hierarchy
		"get_industry_sectors", "get_industry_sector", "get_industry_divisions", "get_industry_division",
		"get_industry_groups", "get_industry_group", "get_industry_classes", "get_industry_class",
		"get_industry_subclasses", "get_industry_subclass",
		// breakdowns by level
		"get_total_job_count_by_level", "get_employment_sector_by_level", "get_experience_by_level",
		"get_province_by_level", "get_education_by_level", "get_formality_by_level",
		"get_gender_by_level", "get_vocational_education_by_level", "get_remote_onsite_by_level",
		"get_job_type_by_level", "get_hierarchy_children",
		// analysis
		"get_all_skills_by_occupation", "get_top_hiring_employers_by_occupation", "get_vacancy_trend",
		"get_vacancy_total", "get_occupations_by_date_range", "get_industries_by_date_range",
		// newspaper upload
		"select_manual_source", "submit_newspaper_vacancies",
	}

	server := New(context.Background(), service.NewServices(repositories.NewStore(nil)), nil)
	session := connectInMemory(t, server, nil)

	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	byName := make(map[string]*mcp.Tool, len(listed.Tools))
	for _, tool := range listed.Tools {
		byName[tool.Name] = tool
	}

	for _, name := range wantTools {
		t.Run(name, func(t *testing.T) {
			tool, ok := byName[name]
			if !ok {
				t.Fatalf("tool %q is not registered", name)
			}
			if strings.TrimSpace(tool.Description) == "" {
				t.Errorf("tool %q has no description, so the model cannot tell when to use it", name)
			}
			if tool.InputSchema == nil {
				t.Errorf("tool %q has no input schema", name)
			}
		})
	}

	// A tool nobody listed above was added without a test; make that visible.
	for name := range byName {
		if !slices.Contains(wantTools, name) {
			t.Errorf("tool %q is registered but missing from wantTools", name)
		}
	}
}

// TestNew_Instructions checks the instructions sent on initialize mention the
// tools the workflow depends on, so renaming a tool cannot silently leave the
// instructions pointing at a name that no longer exists.
func TestNew_Instructions(t *testing.T) {
	server := New(context.Background(), service.NewServices(repositories.NewStore(nil)), nil)
	session := connectInMemory(t, server, nil)

	instructions := session.InitializeResult().Instructions
	for _, name := range []string{
		"get_hierarchy_children", "get_major_groups", "get_industry_sectors",
		"select_manual_source", "submit_newspaper_vacancies",
	} {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(instructions, name) {
				t.Errorf("instructions do not mention %q", name)
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

func TestServeProtectedResourceMetadata(t *testing.T) {
	tests := []struct {
		name string
		meta Metadata
	}{
		{"local", Metadata{ResourceID: "http://localhost:9090/mcp", IssuerURL: "http://localhost:8090", PublicBaseURL: "http://localhost:9090"}},
		{"deployed", Metadata{ResourceID: "https://mcp.example.com/mcp", IssuerURL: "https://auth.example.com", PublicBaseURL: "https://mcp.example.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			serveProtectedResourceMetadata(tt.meta)(w, httptest.NewRequest(http.MethodGet, protectedResPath, nil))

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			var got struct {
				Resource             string   `json:"resource"`
				AuthorizationServers []string `json:"authorization_servers"`
				ScopesSupported      []string `json:"scopes_supported"`
				BearerMethods        []string `json:"bearer_methods_supported"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}
			if got.Resource != tt.meta.ResourceID {
				t.Errorf("resource = %q, want %q", got.Resource, tt.meta.ResourceID)
			}
			if !slices.Equal(got.AuthorizationServers, []string{tt.meta.IssuerURL}) {
				t.Errorf("authorization_servers = %v, want [%s]", got.AuthorizationServers, tt.meta.IssuerURL)
			}
			if !slices.Contains(got.ScopesSupported, submitNewspaperVacanciesScope) {
				t.Errorf("scopes_supported = %v, want it to include %q", got.ScopesSupported, submitNewspaperVacanciesScope)
			}
			if !slices.Equal(got.BearerMethods, []string{"header"}) {
				t.Errorf("bearer_methods_supported = %v, want [header]", got.BearerMethods)
			}
		})
	}
}
