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

// serverInstructions is sent to every MCP client on initialize. It explains
// how the read tools relate to each other, and carries the newspaper-upload
// workflow so the client shows the extracted jobs in the chat and waits for
// the user's go-ahead before anything is saved.
const serverInstructions = labourMarketInstructions + "\n\n" + newspaperInstructions

const labourMarketInstructions = `Labour market data tools - how they fit together:
- Reference lists (get_experiences, get_provinces, get_job_types, get_employment_sectors, get_education_levels, get_formalities, get_genders, get_vocational_educations) only list the available values, with no counts. For vacancy counts by one of those dimensions use the matching get_*_by_level tool.
- Two classification hierarchies exist: 'occupation' (major-group > sub-major-group > minor-group > unit-group > occupation-group) and 'industry' (industry-sector > industry-division > industry-group > industry-class > industry-subclass). A level must belong to the standard it is used with. Use get_major_groups / get_industry_sectors to get top-level ids, and get_hierarchy_children to go one level down (the leaf levels have no children).
- To find a specific leaf (occupation group or industry subclass), drill down with get_hierarchy_children rather than paging through get_occupation_groups / get_industry_subclasses.
- get_occupations_by_date_range / get_industries_by_date_range give one count per top-level group, for an overview or to compare groups. To break down one specific group by province, gender, experience and so on, use the get_*_by_level tools with that group's level and id.
- get_all_skills_by_occupation and get_top_hiring_employers_by_occupation are occupation-only. get_vacancy_trend is national and not scoped to a level.
- Paginated tools return 'total'; the next page starts at offset + limit, and you stop when offset >= total. Omit limit to use the default.`

const newspaperInstructions = `Newspaper vacancy upload workflow. When the user uploads a newspaper image:
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
	// These are reference lists only (no counts). Each description points to
	// the matching get_*_by_level tool for vacancy counts, and uses "e.g."
	// examples rather than full value lists so they don't go stale when the
	// database changes.
	registerNoArgTool(server, "get_experiences",
		"List all experience levels (reference list only, no counts). For vacancy counts by experience level, use get_experience_by_level.",
		plainGetAll(svcs.Experiences.GetAll))

	registerNoArgTool(server, "get_provinces",
		"List all Sri Lankan provinces used for geo-tagging job posts (reference list only, no counts). For vacancy counts by province, use get_province_by_level.",
		plainGetAll(svcs.GeoData.GetAll))

	registerNoArgTool(server, "get_job_types",
		"List all job types, e.g. Full Time or Part Time (reference list only, no counts). For vacancy counts by job type, use get_job_type_by_level.",
		plainGetAll(svcs.JobTypes.GetAll))

	registerNoArgTool(server, "get_employment_sectors",
		"List all employment sectors, e.g. Government or Private (reference list only, no counts). For vacancy counts by employment sector, use get_employment_sector_by_level.",
		plainGetAll(svcs.EmploymentSectors.GetAll))

	registerNoArgTool(server, "get_education_levels",
		"List all education level categories (reference list only, no counts). For vacancy counts by education level, use get_education_by_level.",
		plainGetAll(svcs.EducationLevels.GetAll))

	registerNoArgTool(server, "get_formalities",
		"List all formality categories, i.e. formal vs informal sector (reference list only, no counts). For vacancy counts by formality, use get_formality_by_level.",
		plainGetAll(svcs.Formalities.GetAll))

	registerNoArgTool(server, "get_genders",
		"List all gender categories used in job postings (reference list only, no counts). For vacancy counts by gender, use get_gender_by_level.",
		plainGetAll(svcs.Genders.GetAll))

	registerNoArgTool(server, "get_vocational_educations",
		"List all vocational education (NVQ) levels (reference list only, no counts). For vacancy counts by vocational education level, use get_vocational_education_by_level.",
		plainGetAll(svcs.VocationalEducations.GetAll))
}
