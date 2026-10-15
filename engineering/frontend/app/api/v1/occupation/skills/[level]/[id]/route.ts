import { NextRequest, NextResponse } from "next/server";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// GET /api/v1/occupation/skills/:level/:id?from-date=&to-date=&limit=&offset=
// Fans out to top-15-skills, all-skills (paginated), and top-hiring-employers
// for the given occupation level/id, and returns one combined JSON response.
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ level: string; id: string }> }
) {
  const { level, id } = await params;
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");
  const limit = searchParams.get("limit");
  const offset = searchParams.get("offset");

  const dateParams = new URLSearchParams();
  if (fromDate) dateParams.set("from-date", fromDate);
  if (toDate) dateParams.set("to-date", toDate);
  const dateQueryString = dateParams.toString();

  const allSkillsParams = new URLSearchParams(dateParams);
  if (limit) allSkillsParams.set("limit", limit);
  if (offset) allSkillsParams.set("offset", offset);
  const allSkillsQueryString = allSkillsParams.toString();

  console.log("[occupation-skills] fetching from Go:", { level, id, fromDate, toDate, limit, offset });

  const endpoints = [
    { key: "top_15_skills", url: `${GO_API}/occupation/${level}/${id}/top-15-skills?${dateQueryString}` },
    { key: "all_skills", url: `${GO_API}/occupation/${level}/${id}/all-skills?${allSkillsQueryString}` },
    { key: "top_hiring_employers", url: `${GO_API}/occupation/${level}/${id}/top-hiring-employers?${dateQueryString}` },
  ];

  try {
    const responses = await Promise.all(endpoints.map(({ url }) => fetch(url)));

    for (let i = 0; i < responses.length; i++) {
      if (!responses[i].ok) {
        const errorBody = await responses[i].json().catch(() => ({
          error: `Go backend returned an error from ${endpoints[i].key}`,
        }));
        console.error(`[occupation-skills] Go error from ${endpoints[i].key}:`, errorBody);
        return NextResponse.json(errorBody, { status: responses[i].status });
      }
    }

    const results = await Promise.all(responses.map((res) => res.json()));

    const combined = endpoints.reduce((acc, { key }, i) => {
      acc[key] = results[i];
      return acc;
    }, {} as Record<string, unknown>);

    return NextResponse.json({
      level,
      id,
      from_date: fromDate,
      to_date: toDate,
      ...combined,
    });
  } catch (error) {
    console.error("[occupation-skills] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch occupation skills data",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}