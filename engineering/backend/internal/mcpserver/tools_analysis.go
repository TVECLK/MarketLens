package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"marketlens-go-backend/internal/service"
)

func registerAnalysisTools(server *mcp.Server, svcs *service.Services) {
	// Occupation skills (occupation only - no industry equivalent exists)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_skills_by_occupation",
		Description: "Get all skills (paginated) for a given occupation hierarchy level and id, within a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in struct {
		Level    string `json:"level" jsonschema:"occupation hierarchy level: 'major-group', 'sub-major-group', 'minor-group', 'unit-group' or 'occupation-group'"`
		ID       uint   `json:"id" jsonschema:"the numeric id at that level"`
		FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD"`
		ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD"`
		Limit    int    `json:"limit,omitempty" jsonschema:"max rows to return, default 20, max 100"`
		Offset   int    `json:"offset,omitempty" jsonschema:"rows to skip, default 0"`
	}) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		items, total, err := svcs.Analytics.GetAllSkillsByOccupationLevel(ctx, in.Level, in.ID, from, to, in.Limit, in.Offset)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": items, "total": total, "limit": svcs.Analytics.SkillsPageLimit(in.Limit), "offset": max(in.Offset, 0)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_top_hiring_employers_by_occupation",
		Description: "Get the top hiring employers for a given occupation hierarchy level and id, within a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in struct {
		Level    string `json:"level" jsonschema:"occupation hierarchy level: 'major-group', 'sub-major-group', 'minor-group', 'unit-group' or 'occupation-group'"`
		ID       uint   `json:"id" jsonschema:"the numeric id at that level"`
		FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD"`
		ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD"`
		Limit    int    `json:"limit,omitempty" jsonschema:"max rows to return, default 5, max 100"`
	}) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		out, err := svcs.Analytics.GetHiringEmployersByOccupationLevel(ctx, in.Level, in.ID, from, to, in.Limit)
		if err != nil {
			return nil, nil, err
		}
		return nil, wrapIfList(out), nil
	})

	// Adaptive vacancy trend + total (national, not level-scoped)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_vacancy_trend",
		Description: "Get the national vacancy trend over a date range. Automatically buckets daily " +
			"for ranges under 60 days, or monthly for longer ranges.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in dateRangeInput) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		data, granularity, err := svcs.Analytics.GetVacancyTrend(ctx, from, to)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"granularity": granularity, "data": data}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_vacancy_total",
		Description: "Get the total national vacancy count for a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in dateRangeInput) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		total, err := svcs.Analytics.GetTotalVacancyCount(ctx, from, to)
		return nil, map[string]any{"total_vacancies": total}, err
	})

	// Date-range occupation/industry counts

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_occupations_by_date_range",
		Description: "Get vacancy counts grouped by occupation major group for a date range - one count per " +
			"top-level group. Use this for a top-level overview or to compare groups against each other " +
			"(e.g. which occupations are hiring most). To break down one specific group (by province, " +
			"gender, etc.), use the get_*_by_level tools; to go one level deeper, use get_hierarchy_children.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in dateRangeInput) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		out, err := svcs.Analytics.GetJobCountByDateRange(ctx, "occupation", from, to)
		if err != nil {
			return nil, nil, err
		}
		return nil, wrapIfList(out), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_industries_by_date_range",
		Description: "Get vacancy counts grouped by industry sector for a date range - one count per " +
			"top-level sector. Use this for a top-level overview or to compare sectors against each other " +
			"(e.g. which industries are hiring most). To break down one specific sector (by province, " +
			"gender, etc.), use the get_*_by_level tools; to go one level deeper, use get_hierarchy_children.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in dateRangeInput) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		out, err := svcs.Analytics.GetJobCountByDateRange(ctx, "industry", from, to)
		if err != nil {
			return nil, nil, err
		}
		return nil, wrapIfList(out), nil
	})
}
