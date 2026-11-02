import { NextRequest, NextResponse } from "next/server";
import { fetchSections } from "@/lib/bff-aggregate";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// GET /api/v1/occupation/skills/:level/:id?from-date=&to-date=&limit=&offset=
// Fans out to top-15-skills, all-skills (paginated), and top-hiring-employers
// for the given occupation level/id, and returns one combined JSON response.
// A section that fails comes back as `null` with its reason under `errors`, so
// the sections that worked can still be shown. Only when every section fails
// is the Go error (status and body) forwarded as-is.
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

  const { data, errors, allFailed, firstError } = await fetchSections([
    { key: "top_15_skills", url: `${GO_API}/occupation/${level}/${id}/top-15-skills?${dateQueryString}` },
    { key: "all_skills", url: `${GO_API}/occupation/${level}/${id}/all-skills?${allSkillsQueryString}` },
    { key: "top_hiring_employers", url: `${GO_API}/occupation/${level}/${id}/top-hiring-employers?${dateQueryString}` },
  ]);

  if (allFailed && firstError) {
    return NextResponse.json(firstError.body, { status: firstError.status });
  }

  return NextResponse.json({
    level,
    id,
    from_date: fromDate,
    to_date: toDate,
    ...data,
    errors,
  });
}
