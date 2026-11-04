import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from "@/lib/token";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get all major groups - currently-active by default, or active during a
// given date range if from-date and to-date query params are provided
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);

  const fromDate = searchParams.get("from-date");
  const toDate = searchParams.get("to-date");

  const backendParams = new URLSearchParams();
  if (fromDate) backendParams.set("from-date", fromDate);
  if (toDate) backendParams.set("to-date", toDate);
  const queryString = backendParams.toString();

  const backendUrl = `${BACKEND_API}/major-groups${queryString ? `?${queryString}` : ""}`;

  console.log("[major-groups] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[major-groups] backend error body:", errorText);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[major-groups] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch major groups",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new major group
export async function POST(request: NextRequest) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const backendUrl = `${BACKEND_API}/major-groups`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[major-groups] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[major-groups] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create major group",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}