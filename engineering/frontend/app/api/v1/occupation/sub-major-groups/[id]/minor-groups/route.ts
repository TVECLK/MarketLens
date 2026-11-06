import { NextRequest, NextResponse } from "next/server";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// GET /api/occupation/sub-major-groups/:id/minor-groups
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const backendParams = new URLSearchParams();
  if (fromDate) backendParams.set("from-date", fromDate);
  if (toDate) backendParams.set("to-date", toDate);
  const queryString = backendParams.toString();

  const backendUrl = `${BACKEND_API}/sub-major-groups/${id}/minor-groups${queryString ? `?${queryString}` : ""}`;

  console.log("[sub-major-groups-minor-groups] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);
    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[sub-major-groups-minor-groups] backend error body:", errorBody);
      return NextResponse.json(errorBody ?? { error: "Backend returned an error" }, { status: response.status });
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[sub-major-groups-minor-groups] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch minor groups",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}