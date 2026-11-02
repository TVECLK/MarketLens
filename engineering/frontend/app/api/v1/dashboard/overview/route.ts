import { NextRequest, NextResponse } from "next/server";
import { fetchSections } from "@/lib/bff-aggregate";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// GET /api/v1/dashboard/overview?from-date=YYYY-MM-DD&to-date=YYYY-MM-DD
// Fans out to the Go dashboard endpoints and returns one combined response.
// A section that fails comes back as `null` with its reason under `errors`, so
// the sections that worked can still be shown. Only when every section fails
// is the Go error (status and body) forwarded as-is, e.g. 400 for bad dates.
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const goParams = new URLSearchParams();
  if (fromDate) goParams.set("from-date", fromDate);
  if (toDate) goParams.set("to-date", toDate);
  const queryString = goParams.toString();

  console.log("[dashboard-overview] fetching from Go:", { fromDate, toDate });

  const { data, errors, allFailed, firstError } = await fetchSections([
    { key: "vacancy_trend", url: `${GO_API}/vacancy-trend?${queryString}` },
    { key: "vacancy_total", url: `${GO_API}/vacancy-total?${queryString}` },
    { key: "by_occupation", url: `${GO_API}/occupations/by-date-range?${queryString}` },
    { key: "by_industry", url: `${GO_API}/industries/by-date-range?${queryString}` },
  ]);

  if (allFailed && firstError) {
    return NextResponse.json(firstError.body, { status: firstError.status });
  }

  return NextResponse.json({ ...data, errors });
}
