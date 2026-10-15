import { NextRequest, NextResponse } from "next/server";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// GET /api/dashboard/overview?from-date=YYYY-MM-DD&to-date=YYYY-MM-DD
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const goParams = new URLSearchParams();
  if (fromDate) goParams.set("from-date", fromDate);
  if (toDate) goParams.set("to-date", toDate);
  const queryString = goParams.toString();

  console.log("[dashboard-overview] fetching from Go:", { fromDate, toDate });

  try {
    const [
      vacancyTrendRes,
      vacancyTotalRes,
      occupationsByDateRangeRes,
      industriesByDateRangeRes,
    ] = await Promise.all([
      fetch(`${GO_API}/vacancy-trend?${queryString}`),
      fetch(`${GO_API}/vacancy-total?${queryString}`),
      fetch(`${GO_API}/occupations/by-date-range?${queryString}`),
      fetch(`${GO_API}/industries/by-date-range?${queryString}`),
    ]);

    const responses = [
      { name: "vacancy-trend", res: vacancyTrendRes },
      { name: "vacancy-total", res: vacancyTotalRes },
      { name: "occupations/by-date-range", res: occupationsByDateRangeRes },
      { name: "industries/by-date-range", res: industriesByDateRangeRes },
    ];

    // If any Go endpoint returned an error (400 for bad dates, 500 for a real
    // failure, etc.), forward its exact status and body straight through —
    // Go already built the right error shape and status code for this.
    for (const { name, res } of responses) {
      if (!res.ok) {
        const errorBody = await res.json().catch(() => ({ error: `Go backend returned an error from ${name}` }));
        return NextResponse.json(errorBody, { status: res.status });
      }
    }

    const [vacancyTrend, vacancyTotal, occupationsByDateRange, industriesByDateRange] =
      await Promise.all(responses.map(({ res }) => res.json()));

    return NextResponse.json({
      vacancy_trend: vacancyTrend,
      vacancy_total: vacancyTotal,
      by_occupation: occupationsByDateRange,
      by_industry: industriesByDateRange,
    });
  } catch (error) {
    console.error("[dashboard-overview] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch dashboard overview",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}