import { NextRequest, NextResponse } from "next/server";
import { fetchSections } from "@/lib/bff-aggregate";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// GET /api/v1/dashboard/overview?from-date=YYYY-MM-DD&to-date=YYYY-MM-DD
// Fans out to the backend dashboard endpoints and returns one combined response.
// A section that fails comes back as `null` with its reason under `errors`, so
// the sections that worked can still be shown. Only when every section fails
// is the backend error (status and body) forwarded as-is, e.g. 400 for bad dates.
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const backendParams = new URLSearchParams();
  if (fromDate) backendParams.set("from-date", fromDate);
  if (toDate) backendParams.set("to-date", toDate);
  const queryString = backendParams.toString();

  console.log("[dashboard-overview] fetching from backend:", { fromDate, toDate });

  const { data, errors, allFailed, firstError } = await fetchSections([
    { key: "vacancy_trend", url: `${BACKEND_API}/vacancy-trend?${queryString}` },
    { key: "vacancy_total", url: `${BACKEND_API}/total-vacancy?${queryString}` },
    { key: "by_occupation", url: `${BACKEND_API}/analytics/occupation/vacancy-summary?${queryString}` },
    { key: "by_industry", url: `${BACKEND_API}/analytics/industry/vacancy-summary?${queryString}` },
  ]);

  if (allFailed && firstError) {
    return NextResponse.json(firstError.body, { status: firstError.status });
  }

  return NextResponse.json({ ...data, errors });
}
