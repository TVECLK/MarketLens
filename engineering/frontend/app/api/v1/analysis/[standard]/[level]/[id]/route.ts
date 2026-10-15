import { NextRequest, NextResponse } from "next/server";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

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
  "remote-onsite-hybrid",
  "job-type",
] as const;

// GET /api/analysis/:standard/:level/:id?from-date=&to-date=
// Fans out to all 11 Go analysis endpoints for the given standard/level/id,
// in parallel, and returns one combined JSON response.
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ standard: string; level: string; id: string }> }
) {
  const { standard, level, id } = await params;
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const goParams = new URLSearchParams();
  if (fromDate) goParams.set("from-date", fromDate);
  if (toDate) goParams.set("to-date", toDate);
  const queryString = goParams.toString();

  console.log("[analysis] fetching from Go:", { standard, level, id, fromDate, toDate });

  try {
    const responses = await Promise.all(
      ANALYSIS_TYPES.map((type) =>
        fetch(`${GO_API}/${standard}/${level}/${id}/${type}${queryString ? `?${queryString}` : ""}`)
      )
    );

    for (let i = 0; i < responses.length; i++) {
      const type = ANALYSIS_TYPES[i];
      if (!responses[i].ok && type !== "children") {
        const errorBody = await responses[i].json().catch(() => ({
          error: `Go backend returned an error from ${type}`,
        }));
        console.error(`[analysis] Go error from ${type}:`, errorBody);
        return NextResponse.json(errorBody, { status: responses[i].status });
      }
    }

    const results = await Promise.all(
      responses.map(async (res, i) => {
        if (!res.ok && ANALYSIS_TYPES[i] === "children") {
          return null; 
        }
        return res.json();
      })
    );

    const combined = ANALYSIS_TYPES.reduce((acc, type, i) => {
      acc[type.replace(/-/g, "_")] = results[i];
      return acc;
    }, {} as Record<string, unknown>);

    return NextResponse.json({
      standard,
      level,
      id,
      from_date: fromDate,
      to_date: toDate,
      ...combined,
    });
  } catch (error) {
    console.error("[analysis] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch analysis data",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}