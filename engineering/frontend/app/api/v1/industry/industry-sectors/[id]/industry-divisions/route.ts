import { NextRequest, NextResponse } from "next/server";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// GET /api/industry/industry-sectors/:id/industry-divisions
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const goParams = new URLSearchParams();
  if (fromDate) goParams.set("from-date", fromDate);
  if (toDate) goParams.set("to-date", toDate);
  const queryString = goParams.toString();

  const goUrl = `${GO_API}/industry-sectors/${id}/industry-divisions${queryString ? `?${queryString}` : ""}`;

  console.log("[industry-sectors-industry-divisions] fetching from Go:", goUrl);

  try {
    const response = await fetch(goUrl);
    const data = await response.json();

    if (!response.ok) {
      console.error("[industry-sectors-industry-divisions] Go error body:", data);
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-sectors-industry-divisions] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch industry divisions",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}