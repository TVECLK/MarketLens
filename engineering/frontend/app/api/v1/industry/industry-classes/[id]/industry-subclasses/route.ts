import { NextRequest, NextResponse } from "next/server";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// GET /api/industry/industry-classes/:id/industry-subclasses
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

  const backendUrl = `${BACKEND_API}/industry-classes/${id}/industry-subclasses${queryString ? `?${queryString}` : ""}`;

  console.log("[industry-classes-industry-subclasses] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);
    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[industry-classes-industry-subclasses] backend error body:", errorBody);
      return NextResponse.json(errorBody ?? { error: "Backend returned an error" }, { status: response.status });
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-classes-industry-subclasses] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch industry subclasses",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}