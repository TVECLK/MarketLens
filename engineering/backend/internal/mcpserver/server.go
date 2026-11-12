package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"marketlens-go-backend/internal/auth"
	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/service"
)

const (
	protectedResPath = "/.well-known/oauth-protected-resource/mcp"
)

type contextKey string

// Metadata holds the values needed for the OAuth protected-resource
// metadata endpoint and the WWW-Authenticate challenge header — loaded
// by the caller (see cmd/mcp/main.go) and passed in explicitly, not read
// from the environment here.
type Metadata struct {
	ResourceID    string
	IssuerURL     string
	PublicBaseURL string
}

func serveProtectedResourceMetadata(meta Metadata) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"resource":                 meta.ResourceID,
			"authorization_servers":    []string{meta.IssuerURL},
			"scopes_supported":         []string{"submit-newspaper-vacancies"},
			"bearer_methods_supported": []string{"header"},
		})
	}
}

const scopesKey contextKey = "scopes"

// serverInstructions is sent to every MCP client on initialize. It carries
// the newspaper-upload workflow so the client shows the extracted jobs in
// the chat and waits for the user's go-ahead before anything is saved.
const serverInstructions = `Newspaper vacancy upload workflow. When the user uploads a newspaper image:
1. Read the image yourself and extract every job ad: employer, job role, location, description.
2. Call select_manual_source so the user chooses which newspaper the image came from. The user may only choose a source it offers. If it returns status choose_in_chat, list those names and ask the user to pick one. If the user says none match, or there are no sources, tell them: "If there is no matching source, you should add source from the dashboard and then try again." Do not continue without a selected source.
3. Show the extracted jobs in the chat as a numbered table (employer, role, location, description) together with the chosen source.
4. Call submit_newspaper_vacancies. It shows the user a Yes / No / Edit form and saves only on Yes.
   - status cancelled (user chose No): stop. Nothing was saved.
   - status edits_requested (user chose Edit): apply the returned edits to the jobs (if none were given, ask the user in the chat what to change), show the updated numbered table in the chat, then call submit_newspaper_vacancies again with the updated jobs so the user confirms again.
   - saved: tell the user how many were submitted and list any that failed.
5. Only if the tool says the client cannot show a confirmation form: ask the user in the chat, and call it again with user_confirmed=true only after their explicit yes.
Never save without the user's explicit Yes, and never set user_confirmed on your own.`

func requireValidToken(verifier *auth.Verifier, publicBaseURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("WWW-Authenticate",
				`Bearer resource_metadata="`+publicBaseURL+protectedResPath+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := verifier.VerifyToken(rawToken)
		if err != nil {
			log.Printf("mcp token verification failed: %v", err)
			if errors.Is(err, auth.ErrJWKSUnavailable) {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), scopesKey, strings.Fields(claims.Scope))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// New builds the MCP server and registers every read-only (GET-equivalent)
// tool against the given services.
func New(svcs *service.Services, ingestionSvc *ingestion.IngestionService) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "marketlens-mcp",
		Version: "v1.0.0",
	}, &mcp.ServerOptions{Instructions: serverInstructions})

	registerLookupTools(server, svcs)
	registerHierarchyTools(server, svcs)
	registerAnalysisTools(server, svcs)
	registerManualUploadTools(server, svcs, ingestionSvc)

	return server
}

// StartHTTP serves the given MCP server over Streamable HTTP at addr
// (e.g. ":9090"), authenticating every request against verifier.
func StartHTTP(server *mcp.Server, verifier *auth.Verifier, meta Metadata, addr string) error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)

	mux := http.NewServeMux()
	mux.HandleFunc(protectedResPath, serveProtectedResourceMetadata(meta))
	mux.Handle("/mcp", requireValidToken(verifier, meta.PublicBaseURL, handler))

	return http.ListenAndServe(addr, mux)
}

func wrapIfList(v any) any {
	if v == nil {
		return v
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return map[string]any{"items": v}
	case reflect.Map, reflect.Struct, reflect.Ptr:
		return v
	default:
		return map[string]any{"value": v}
	}
}

// emptyInput is used for tools that take no parameters at all.
type emptyInput struct{}

// registerNoArgTool registers a tool that takes no input and returns
// whatever the given repository call returns.
func registerNoArgTool[T any](server *mcp.Server, name, description string, fn func(ctx context.Context) (T, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
			out, err := fn(ctx)
			if err != nil {
				return nil, nil, err
			}
			return nil, wrapIfList(out), nil
		},
	)
}

// idInput is used for tools scoped to a single numeric id (e.g. an
// industry_sector id, a major_group id).
type idInput struct {
	ID uint `json:"id" jsonschema:"the numeric id to look up"`
}

// registerIDTool registers a tool that takes a single "id" parameter.
func registerIDTool[T any](server *mcp.Server, name, description string, fn func(ctx context.Context, id uint) (T, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description},
		func(ctx context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, any, error) {
			out, err := fn(ctx, in.ID)
			if err != nil {
				return nil, nil, err
			}
			return nil, wrapIfList(out), nil
		},
	)
}

// dateRangeInput is used for tools scoped to a from/to date window.
type dateRangeInput struct {
	FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD"`
	ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD"`
}

// idAndYearInput is used for the occupation-scoped year-filtered tools
// (by-formality, by-gender, top-job-roles).
type idAndYearInput struct {
	ID   uint `json:"id" jsonschema:"the numeric id to look up"`
	Year int  `json:"year" jsonschema:"the calendar year, e.g. 2026"`
}

// idAndDateRangeInput is used for tools scoped to both a single id and a
// from/to date window (e.g. industry-scoped year analytics).
type idAndDateRangeInput struct {
	ID       uint   `json:"id" jsonschema:"the numeric id to look up"`
	FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD"`
	ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD"`
}

// plainGetAll adapts a CRUDService's unified GetAll (limit/offset/date-range
// all optional) to the no-arg shape registerNoArgTool expects, for the MCP
// lookup tools that always want the full, unfiltered list.
func plainGetAll[T any](getAll func(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, int64, error)) func(ctx context.Context) ([]T, error) {
	return func(ctx context.Context) ([]T, error) {
		items, _, err := getAll(ctx, 0, 0, nil, nil)
		return items, err
	}
}

func registerLookupTools(server *mcp.Server, svcs *service.Services) {
	registerNoArgTool(server, "get_experiences", "List all experience levels.",
		plainGetAll(svcs.Experiences.GetAll))

	registerNoArgTool(server, "get_provinces", "List all Sri Lankan provinces used for geo-tagging job posts.",
		plainGetAll(svcs.GeoData.GetAll))

	registerNoArgTool(server, "get_job_types", "List all job types (Full Time, Part Time, Contract, Internship).",
		plainGetAll(svcs.JobTypes.GetAll))

	registerNoArgTool(server, "get_employment_sectors", "List all employment sectors (Government, Private, NGO, etc.).",
		plainGetAll(svcs.EmploymentSectors.GetAll))

	registerNoArgTool(server, "get_education_levels", "List all education level categories.",
		plainGetAll(svcs.EducationLevels.GetAll))

	registerNoArgTool(server, "get_formalities", "List all formality categories (Formal / Informal sector).",
		plainGetAll(svcs.Formalities.GetAll))

	registerNoArgTool(server, "get_genders", "List all gender categories used in job postings.",
		plainGetAll(svcs.Genders.GetAll))

	registerNoArgTool(server, "get_vocational_educations", "List all vocational education (NVQ) levels.",
		plainGetAll(svcs.VocationalEducations.GetAll))
}
