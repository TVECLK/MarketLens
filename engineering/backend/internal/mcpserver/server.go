package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"marketlens-go-backend/internal/auth"
	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/service"
)

const (
	protectedResPath = "/.well-known/oauth-protected-resource/mcp"
)

// tokenInfoCtxKey carries the verified token's info from requireValidToken to
// bridgeTokenInfo within one HTTP request.
type tokenInfoCtxKey struct{}

// errMissingSubject rejects a token with no "sub" claim: the SDK only binds a
// session to its creator when TokenInfo.UserID is set, so an empty one would
// silently switch the session-hijacking check off.
var errMissingSubject = errors.New("token has no subject")

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

// serverInstructions is sent to every MCP client on initialize. It explains
// how the read tools relate to each other, and carries the newspaper-upload
// workflow so the client shows the extracted jobs in the chat and waits for
// the user's go-ahead before anything is saved.
const serverInstructions = labourMarketInstructions + "\n\n" + newspaperInstructions

const labourMarketInstructions = `Labour market data tools - how they fit together:
- All counts are numbers of job vacancies (from job posts), not numbers of workers or employment. Describe them as vacancies.
- Statistics tools need from_date and to_date (YYYY-MM-DD). If the user does not give a period, use January 1 of the current year to today and tell them which period you used.
- Reference lists (get_experiences, get_provinces, get_job_types, get_employment_sectors, get_education_levels, get_formalities, get_genders, get_vocational_educations) only list the available values, with no counts. For vacancy counts by one of those dimensions use the matching get_*_by_level tool.
- Two classification hierarchies exist: 'occupation' (major-group > sub-major-group > minor-group > unit-group > occupation-group) and 'industry' (industry-sector > industry-division > industry-group > industry-class > industry-subclass). A level must belong to the standard it is used with. Use get_major_groups / get_industry_sectors to get top-level ids, and get_hierarchy_children to go one level down (the leaf levels have no children).
- To find a specific leaf (occupation group or industry subclass), drill down with get_hierarchy_children rather than paging through get_occupation_groups / get_industry_subclasses.
- get_occupations_by_date_range / get_industries_by_date_range give one count per top-level group, for an overview or to compare groups. To break down one specific group by province, gender, experience and so on, use the get_*_by_level tools with that group's level and id.
- get_all_skills_by_occupation and get_top_hiring_employers_by_occupation are occupation-only. get_vacancy_trend is national and not scoped to a level.
- Paginated tools return 'total'; the next page starts at offset + limit, and you stop when offset >= total. Omit limit to use the default.`

const newspaperInstructions = `Newspaper vacancy upload workflow. When the user uploads a newspaper image:
1. Read the image yourself and extract every job ad: employer, job role, location, description.
2. Call select_manual_source so the user chooses which newspaper the image came from. The user may only choose a source it offers. If it returns status choose_in_chat, list those names and ask the user to pick one. If it returns status cancelled (the user closed the select box without choosing), stop and relay its message. If the user says none match, or there are no sources, tell them: "If there is no matching source, you should add a source from the dashboard and then try again." Do not continue without a selected source. If either upload tool fails with a missing-scope error, the user is not allowed to upload newspaper vacancies: stop and tell them they do not have permission.
3. Show the extracted jobs in the chat as a numbered table (employer, role, location, description) together with the chosen source.
4. Call submit_newspaper_vacancies. It shows the user a Yes / No / Edit form and saves only on Yes.
   - status cancelled (user chose No): stop. Nothing was saved.
   - status edits_requested (user chose Edit): apply the returned edits to the jobs (if none were given, or they are unclear or don't match any job, ask the user in the chat what they meant instead of guessing), show the updated numbered table in the chat, then call submit_newspaper_vacancies again with the updated jobs so the user confirms again.
   - status server_restarting (the server began shutting down before the user answered): nothing was saved. Tell the user, then call submit_newspaper_vacancies again with the same jobs after a short wait.
   - On success (the user chose Yes) the result has no status; it has 'submitted' (jobs sent), 'saved_count', 'failed_count' and 'failed'. Tell the user how many were saved (saved_count, which includes jobs that matched an existing duplicate) and list any that failed. Each upload is treated as the source's full current list: earlier open vacancies from the same source that are not in this upload are closed, and 'reconciled_stale_jobs' is how many were closed. If it is above 0, tell the user that many earlier vacancies from the source were closed. If 'reconcile_note' is set, tell the user what it says; the jobs that were saved stay saved. If 'interrupted' is true, the upload was stopped before it finished and no earlier vacancies were closed: tell the user, and call submit_newspaper_vacancies again with the FULL list of jobs (jobs already saved are recognised as duplicates, so nothing is saved twice) - never with only the failed jobs, because closing earlier vacancies would then close the ones already saved.
5. Only if the tool says the client cannot show a confirmation form: ask the user in the chat, and call it again with user_confirmed=true only after their explicit yes.
Never save without the user's explicit Yes, and never set user_confirmed on your own.`

// verifyFunc verifies a raw bearer token and returns the token info the MCP
// SDK needs. Errors wrapping auth.ErrJWKSUnavailable become 503; any other
// error becomes 401.
type verifyFunc func(rawToken string) (*mcpauth.TokenInfo, error)

// newVerifyFunc adapts the JWT verifier. UserID is the token's subject, so the
// SDK can refuse a session id used by a different user.
func newVerifyFunc(verifier *auth.Verifier) verifyFunc {
	return func(rawToken string) (*mcpauth.TokenInfo, error) {
		claims, err := verifier.VerifyToken(rawToken)
		if err != nil {
			return nil, err
		}
		if claims.Subject == "" {
			return nil, errMissingSubject
		}
		if claims.ExpiresAt == nil {
			return nil, errors.New("token has no expiry")
		}
		return &mcpauth.TokenInfo{
			Scopes:     strings.Fields(claims.Scope),
			Expiration: claims.ExpiresAt.Time,
			UserID:     claims.Subject,
		}, nil
	}
}

// bearerChallenge is the WWW-Authenticate value for a 401. The MCP auth spec
// wants the resource metadata URL on every 401 so the client knows to log in
// again. A token that was sent but rejected (invalid or expired) also gets
// error="invalid_token" (RFC 6750); a request with no token gets none.
func bearerChallenge(publicBaseURL string, tokenRejected bool) string {
	challenge := `Bearer `
	if tokenRejected {
		challenge += `error="invalid_token", `
	}
	return challenge + `resource_metadata="` + publicBaseURL + protectedResPath + `"`
}

// requireValidToken verifies the bearer token on every request and keeps the
// token info in the request context for bridgeTokenInfo.
func requireValidToken(verify verifyFunc, publicBaseURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("WWW-Authenticate", bearerChallenge(publicBaseURL, false))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		info, err := verify(rawToken)
		if err != nil {
			log.Printf("mcp token verification failed: %v", err)
			if errors.Is(err, auth.ErrJWKSUnavailable) {
				// Not a login problem, so no challenge: sending clients to log
				// in again would only add load while the key set is down.
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("WWW-Authenticate", bearerChallenge(publicBaseURL, true))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), tokenInfoCtxKey{}, info)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// bridgeTokenInfo hands the already-verified token info to the MCP SDK for
// every request. The SDK then (a) passes the current request's token info to
// tool handlers, so scope checks never use a stale session's token, and (b)
// refuses a session id used by a different user. It does not verify the
// token again, so requireValidToken keeps owning the 401/503 responses.
func bridgeTokenInfo(next http.Handler) http.Handler {
	return mcpauth.RequireBearerToken(
		func(_ context.Context, _ string, r *http.Request) (*mcpauth.TokenInfo, error) {
			info, _ := r.Context().Value(tokenInfoCtxKey{}).(*mcpauth.TokenInfo)
			if info == nil {
				return nil, mcpauth.ErrInvalidToken
			}
			return info, nil
		}, nil)(next)
}

// New builds the MCP server and registers every read-only (GET-equivalent)
// tool against the given services.
//
// shutdown is cancelled when the process starts shutting down. A newspaper
// submission in progress then stops retrying and is cut off after
// drainWorkBudget, so it can close its crawler run and answer before
// shutdownGracePeriod runs out.
func New(shutdown context.Context, svcs *service.Services, ingestionSvc *ingestion.IngestionService) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "marketlens-mcp",
		Version: "v1.0.0",
	}, &mcp.ServerOptions{Instructions: serverInstructions})

	registerLookupTools(server, svcs)
	registerHierarchyTools(server, svcs)
	registerAnalysisTools(server, svcs)
	registerManualUploadTools(server, shutdown, svcs, ingestionSvc)

	return server
}

const (
	// readHeaderTimeout bounds how long a client may take to send its request
	// headers, so slow connections cannot tie up the server (Slowloris). There
	// is deliberately no read or write timeout: tool calls such as a newspaper
	// submission run for minutes, and the SDK holds long-lived SSE streams.
	readHeaderTimeout = 10 * time.Second
	idleTimeout       = 2 * time.Minute

	// sessionTimeout closes an MCP session that has received no request for
	// this long. The SDK default (zero) never closes one, so abandoned
	// sessions would pile up in memory. A client whose session expired gets
	// 404 on its next request and starts a new session, as the MCP spec
	// requires of clients.
	sessionTimeout = 30 * time.Minute

	// shutdownGracePeriod is how long in-flight requests get to finish after
	// SIGTERM. Keep it below the platform's kill deadline (Kubernetes
	// terminationGracePeriodSeconds, 30s by default), or the process is killed
	// before this expires. A newspaper submission is held to drainWorkBudget
	// of it, so it cannot outlast the period.
	shutdownGracePeriod = 25 * time.Second
)

// StartHTTP serves the given MCP server over Streamable HTTP at addr
// (e.g. ":9090"), authenticating every request against verifier. When ctx is
// cancelled (SIGTERM) it stops accepting connections and gives in-flight
// requests shutdownGracePeriod to finish, so a newspaper submission is not
// cut off half-saved with its crawler run still open.
func StartHTTP(ctx context.Context, server *mcp.Server, verifier *auth.Verifier, meta Metadata, addr string) error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{SessionTimeout: sessionTimeout})

	tracker := newSessionTracker()

	mux := http.NewServeMux()
	mux.HandleFunc(protectedResPath, serveProtectedResourceMetadata(meta))
	mux.Handle("/mcp", requireValidToken(newVerifyFunc(verifier), meta.PublicBaseURL, bridgeTokenInfo(tracker.track(handler))))

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}
	return serveUntilDone(ctx, srv, ln, shutdownGracePeriod, func(drainCtx context.Context) {
		closeIdleSessions(drainCtx, server, tracker)
	})
}

// serveUntilDone runs srv on ln until it fails or ctx is cancelled. On cancel
// it shuts down gracefully, waiting up to grace for in-flight requests; if they
// do not finish in time it closes their connections and reports an error.
//
// onShutdown, if set, runs concurrently with the drain, until it returns or
// the grace period ends. It is how the MCP server ends the long-lived SSE
// streams that Shutdown would otherwise wait on.
func serveUntilDone(ctx context.Context, srv *http.Server, ln net.Listener, grace time.Duration, onShutdown func(drainCtx context.Context)) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	log.Printf("MCP server shutting down, waiting up to %s for in-flight requests", grace)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	if onShutdown != nil {
		go onShutdown(shutdownCtx)
	}
	if err := srv.Shutdown(shutdownCtx); err != nil {
		srv.Close()
		return fmt.Errorf("graceful shutdown did not finish, in-flight requests were cut off: %w", err)
	}
	return nil
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
				return nil, nil, toolError(name, err)
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
				return nil, nil, toolError(name, err)
			}
			return nil, wrapIfList(out), nil
		},
	)
}

// dateRangeInput is used for tools scoped to a from/to date window.
type dateRangeInput struct {
	FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD. If the user gave no period, use January 1 of the current year and tell them the period you used"`
	ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD. If the user gave no period, use today"`
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
