import { NextRequest, NextResponse } from "next/server";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// GET /api/industry/industry-groups/:id/industry-classes
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

  const backendUrl = `${BACKEND_API}/industry-groups/${id}/industry-classes${queryString ? `?${queryString}` : ""}`;

  console.log("[industry-groups-industry-classes] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);
    const data = await response.json();

    if (!response.ok) {
      console.error("[industry-groups-industry-classes] backend error body:", data);
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-groups-industry-classes] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch industry classes",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}