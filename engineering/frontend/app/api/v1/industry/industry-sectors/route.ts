import { NextRequest, NextResponse } from "next/server";
import { withAuth } from "@/lib/with-auth";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get all industry sectors - currently-active by default, or active during a
// given date range if from-date and to-date query params are provided
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const backendParams = new URLSearchParams();
  if (fromDate) backendParams.set("from-date", fromDate);
  if (toDate) backendParams.set("to-date", toDate);
  const queryString = backendParams.toString();

  const backendUrl = `${BACKEND_API}/industry-sectors${queryString ? `?${queryString}` : ""}`;

  console.log("[industry-sectors] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[industry-sectors] backend error body:", errorText);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-sectors] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch industry sectors",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new industry sector
export const POST = withAuth(async (request, { token }) => {
  const backendUrl = `${BACKEND_API}/industry-sectors`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[industry-sectors] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[industry-sectors] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create industry sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});