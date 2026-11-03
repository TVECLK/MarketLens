import { NextRequest, NextResponse } from "next/server";
import { fetchSections } from "@/lib/bff-aggregate";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

const ANALYSIS_TYPES = [
  "total-job-count",
  "children",
  "employment-sector",
  "experience",
  "province",
  "education",
  "formality",
  "gender",
  "vocational-education",
  "work-modes",
  "job-type",
] as const;

// GET /api/v1/analysis/:standard/:level/:id?from-date=&to-date=
// Fans out to all 11 backend analysis endpoints for the given standard/level/id,
// in parallel, and returns one combined JSON response. A section that fails
// comes back as `null` with its reason under `errors`, so the sections that
// worked can still be shown. `children` is allowed to fail silently: leaf
// levels have none. Only when every other section fails is the backend error
// (status and body) forwarded as-is.
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ standard: string; level: string; id: string }> }
) {
  const { standard, level, id } = await params;
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const backendParams = new URLSearchParams();
  if (fromDate) backendParams.set("from-date", fromDate);
  if (toDate) backendParams.set("to-date", toDate);
  const queryString = backendParams.toString();

  console.log("[analysis] fetching from backend:", { standard, level, id, fromDate, toDate });

  const { data, errors, allFailed, firstError } = await fetchSections(
    ANALYSIS_TYPES.map((type) => ({
      key: type.replace(/-/g, "_"),
      url: `${BACKEND_API}/${standard}/${level}/${id}/${type}${queryString ? `?${queryString}` : ""}`,
      optional: type === "children",
    })),
  );

  if (allFailed && firstError) {
    return NextResponse.json(firstError.body, { status: firstError.status });
  }

  return NextResponse.json({
    standard,
    level,
    id,
    from_date: fromDate,
    to_date: toDate,
    ...data,
    errors,
  });
}
