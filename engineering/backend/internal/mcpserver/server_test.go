package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"marketlens-go-backend/internal/auth"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const testBaseURL = "https://example.test"

// fakeVerify maps a raw token to token info, standing in for the JWT verifier.
func fakeVerify(raw string) (*mcpauth.TokenInfo, error) {
	exp := time.Now().Add(time.Hour)
	switch raw {
	case "alice-upload":
		return &mcpauth.TokenInfo{UserID: "alice", Scopes: []string{submitNewspaperVacanciesScope}, Expiration: exp}, nil
	case "alice-no-scope":
		return &mcpauth.TokenInfo{UserID: "alice", Scopes: []string{"openid"}, Expiration: exp}, nil
	case "bob-upload":
		return &mcpauth.TokenInfo{UserID: "bob", Scopes: []string{submitNewspaperVacanciesScope}, Expiration: exp}, nil
	case "no-subject":
		return nil, errMissingSubject
	case "jwks-down":
		return nil, auth.ErrJWKSUnavailable
	}
	return nil, errors.New("bad token")
}

func TestRequireValidToken_Responses(t *testing.T) {
	handler := requireValidToken(fakeVerify, testBaseURL, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const metadataURL = testBaseURL + protectedResPath
	noToken := `Bearer resource_metadata="` + metadataURL + `"`
	rejected := `Bearer error="invalid_token", resource_metadata="` + metadataURL + `"`

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantHeader string // exact WWW-Authenticate value; empty means none
	}{
		{"no header", "", http.StatusUnauthorized, noToken},
		{"not bearer", "Basic abc", http.StatusUnauthorized, noToken},
		{"invalid token", "Bearer nope", http.StatusUnauthorized, rejected},
		{"token without subject", "Bearer no-subject", http.StatusUnauthorized, rejected},
		{"jwks unavailable", "Bearer jwks-down", http.StatusServiceUnavailable, ""},
		{"valid token", "Bearer alice-upload", http.StatusOK, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if got := w.Header().Get("WWW-Authenticate"); got != tt.wantHeader {
				t.Errorf("WWW-Authenticate = %q, want %q", got, tt.wantHeader)
			}
		})
	}
}

// tokenSwitcher sets the bearer token on every request, and lets a test change
// the token mid-session the way a refreshed or revoked token would.
type tokenSwitcher struct {
	token atomic.Value
	next  http.RoundTripper
}

func (s *tokenSwitcher) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+s.token.Load().(string))
	return s.next.RoundTrip(r)
}

// TestSession_ScopesPerRequestAndUserBinding runs the real SDK handler behind
// our middleware and checks the two things the old context-based scope check
// got wrong: scopes follow the current request's token, and a session id used
// by a different user is refused.
func TestSession_ScopesPerRequestAndUserBinding(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "scope_check"},
		func(_ context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
			return nil, map[string]any{"has_scope": hasScope(req, submitNewspaperVacanciesScope)}, nil
		})

	sdkHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	srv := httptest.NewServer(requireValidToken(fakeVerify, testBaseURL, bridgeTokenInfo(sdkHandler)))
	t.Cleanup(srv.Close)

	switcher := &tokenSwitcher{next: http.DefaultTransport}
	switcher.token.Store("alice-upload")

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             srv.URL,
		HTTPClient:           &http.Client{Transport: switcher},
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	// Steps run in order against the one open session: each step's token is
	// what the client sends from then on.
	steps := []struct {
		name  string
		token string
		// rawRequest sends a plain HTTP request carrying the session id and
		// checks the status, instead of calling the tool through the client.
		rawRequest bool
		wantScope  bool // tool steps: what scope_check should report
		wantStatus int  // raw steps: the expected HTTP status
	}{
		{name: "token with the scope", token: "alice-upload", wantScope: true},
		{name: "same user, scope since removed", token: "alice-no-scope", wantScope: false},
		{name: "scope granted again", token: "alice-upload", wantScope: true},
		{name: "different user on this session", token: "bob-upload", rawRequest: true, wantStatus: http.StatusForbidden},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			if step.rawRequest {
				// Sends its own Authorization header and does not touch the
				// switcher, so the client transport (used by the cleanup that
				// closes the session) keeps the session owner's token.
				req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(`{"jsonrpc":"2.0","id":99,"method":"tools/list"}`))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Authorization", "Bearer "+step.token)
				req.Header.Set("Mcp-Session-Id", session.ID())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json, text/event-stream")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != step.wantStatus {
					t.Errorf("status = %d, want %d", resp.StatusCode, step.wantStatus)
				}
				return
			}

			switcher.token.Store(step.token)
			res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "scope_check"})
			if err != nil {
				t.Fatalf("call tool: %v", err)
			}
			if got := boolField(t, res, "has_scope"); got != step.wantScope {
				t.Errorf("has_scope = %v, want %v", got, step.wantScope)
			}
		})
	}
}

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

func TestServeUntilDone(t *testing.T) {
	tests := []struct {
		name         string
		sendRequest  bool          // start a request before shutdown begins
		handlerDelay time.Duration // how long that request takes
		grace        time.Duration
		wantErr      bool // serveUntilDone reports in-flight requests were cut off
		wantFinished bool // the in-flight request got its response
	}{
		{name: "idle server stops cleanly", grace: time.Second},
		{name: "in-flight request finishes within the grace period", sendRequest: true, handlerDelay: 200 * time.Millisecond, grace: 5 * time.Second, wantFinished: true},
		{name: "in-flight request outlasts the grace period", sendRequest: true, handlerDelay: 10 * time.Second, grace: 100 * time.Millisecond, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			started := make(chan struct{})
			srv := &http.Server{
				ReadHeaderTimeout: time.Second,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					close(started)
					select {
					case <-time.After(tt.handlerDelay):
						w.WriteHeader(http.StatusOK)
					case <-r.Context().Done(): // connection closed by a forced shutdown
					}
				}),
			}
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- serveUntilDone(ctx, srv, ln, tt.grace, nil) }()

			finished := make(chan bool, 1)
			if tt.sendRequest {
				go func() {
					resp, err := http.Get("http://" + ln.Addr().String())
					if err != nil {
						finished <- false
						return
					}
					resp.Body.Close()
					finished <- resp.StatusCode == http.StatusOK
				}()
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					t.Fatal("request never reached the handler")
				}
			}

			cancel() // what SIGTERM does

			select {
			case err := <-done:
				if (err != nil) != tt.wantErr {
					t.Errorf("serveUntilDone error = %v, wantErr %v", err, tt.wantErr)
				}
			case <-time.After(15 * time.Second):
				t.Fatal("serveUntilDone did not return")
			}

			if tt.sendRequest {
				select {
				case got := <-finished:
					if got != tt.wantFinished {
						t.Errorf("in-flight request finished = %v, want %v", got, tt.wantFinished)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("client never got an outcome for the in-flight request")
				}
			}
		})
	}
}
