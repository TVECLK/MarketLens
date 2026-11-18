package mcpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"marketlens-go-backend/internal/auth"
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

func TestRequireValidToken_Responses(t *testing.T) {
	handler := requireValidToken(fakeVerify, testBaseURL, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantHeader bool
	}{
		{"no header", "", http.StatusUnauthorized, true},
		{"not bearer", "Basic abc", http.StatusUnauthorized, true},
		{"invalid token", "Bearer nope", http.StatusUnauthorized, false},
		{"token without subject", "Bearer no-subject", http.StatusUnauthorized, false},
		{"jwks unavailable", "Bearer jwks-down", http.StatusServiceUnavailable, false},
		{"valid token", "Bearer alice-upload", http.StatusOK, false},
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
			if got := w.Header().Get("WWW-Authenticate") != ""; got != tt.wantHeader {
				t.Errorf("WWW-Authenticate set = %v, want %v", got, tt.wantHeader)
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
			switcher.token.Store(step.token)

			if step.rawRequest {
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

			res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "scope_check"})
			if err != nil {
				t.Fatalf("call tool: %v", err)
			}
			out, _ := res.StructuredContent.(map[string]any)
			if got, _ := out["has_scope"].(bool); got != step.wantScope {
				t.Errorf("has_scope = %v, want %v", got, step.wantScope)
			}
		})
	}
}
