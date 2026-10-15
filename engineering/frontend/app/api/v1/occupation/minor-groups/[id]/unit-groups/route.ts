import { NextRequest, NextResponse } from "next/server";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// GET /api/occupation/minor-groups/:id/unit-groups
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

  const goUrl = `${GO_API}/minor-groups/${id}/unit-groups${queryString ? `?${queryString}` : ""}`;

  console.log("[minor-groups-unit-groups] fetching from Go:", goUrl);

  try {
    const response = await fetch(goUrl);
    const data = await response.json();

    if (!response.ok) {
      console.error("[minor-groups-unit-groups] Go error body:", data);
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[minor-groups-unit-groups] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch unit groups",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}