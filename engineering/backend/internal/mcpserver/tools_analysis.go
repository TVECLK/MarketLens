package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"marketlens-go-backend/internal/service"
)

func registerAnalysisTools(server *mcp.Server, svcs *service.Services) {
	// Occupation skills (occupation only - no industry equivalent exists)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_top_15_skills_by_occupation",
		Description: "Get the top 15 in-demand skills for a given occupation hierarchy level and id, within a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in struct {
		Level    string `json:"level" jsonschema:"occupation hierarchy level, e.g. 'major-group'"`
		ID       uint   `json:"id" jsonschema:"the numeric id at that level"`
		FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD"`
		ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD"`
	}) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		out, err := svcs.Analytics.GetTop15SkillsByOccupationLevel(ctx, in.Level, in.ID, from, to)
		if err != nil {
			return nil, nil, err
		}
		return nil, wrapIfList(out), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_all_skills_by_occupation",
		Description: "Get all skills (paginated) for a given occupation hierarchy level and id, within a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in struct {
		Level    string `json:"level" jsonschema:"occupation hierarchy level, e.g. 'major-group'"`
		ID       uint   `json:"id" jsonschema:"the numeric id at that level"`
		FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD"`
		ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD"`
		Limit    int    `json:"limit" jsonschema:"max rows to return, default 20"`
		Offset   int    `json:"offset" jsonschema:"rows to skip, default 0"`
	}) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 20
		}
		items, total, err := svcs.Analytics.GetAllSkillsByOccupationLevel(ctx, in.Level, in.ID, from, to, limit, in.Offset)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"items": items, "total": total, "limit": limit, "offset": in.Offset}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_top_hiring_employers_by_occupation",
		Description: "Get the top hiring employers for a given occupation hierarchy level and id, within a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in struct {
		Level    string `json:"level" jsonschema:"occupation hierarchy level, e.g. 'major-group'"`
		ID       uint   `json:"id" jsonschema:"the numeric id at that level"`
		FromDate string `json:"from_date" jsonschema:"start date, format YYYY-MM-DD"`
		ToDate   string `json:"to_date" jsonschema:"end date, format YYYY-MM-DD"`
	}) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		out, err := svcs.Analytics.GetTopHiringEmployersByOccupationLevel(ctx, in.Level, in.ID, from, to)
		if err != nil {
			return nil, nil, err
		}
		return nil, wrapIfList(out), nil
	})

	// Adaptive vacancy trend + total (national, not level-scoped)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_vacancy_trend",
		Description: "Get the national vacancy trend over a date range. Automatically buckets daily " +
			"for ranges under 60 days, or monthly for longer ranges - matching GET /vacancy-trend.",
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
		Name:        "get_occupations_by_date_range",
		Description: "Get vacancy counts grouped by occupation major group for a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in dateRangeInput) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		out, err := svcs.Analytics.GetOccupationJobCountByDateRange(ctx, from, to)
		if err != nil {
			return nil, nil, err
		}
		return nil, wrapIfList(out), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_industries_by_date_range",
		Description: "Get vacancy counts grouped by industry sector for a date range.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in dateRangeInput) (*mcp.CallToolResult, any, error) {
		from, to, err := service.ParseDateRange(in.FromDate, in.ToDate)
		if err != nil {
			return nil, nil, err
		}
		out, err := svcs.Analytics.GetIndustryJobCountByDateRange(ctx, from, to)
		if err != nil {
			return nil, nil, err
		}
		return nil, wrapIfList(out), nil
	})
}
