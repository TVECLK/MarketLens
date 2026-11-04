import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from "@/lib/token";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get all occupation groups with pagination
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const limit = searchParams.get("limit");
  const offset = searchParams.get("offset");

  const backendParams = new URLSearchParams();
  if (limit) backendParams.set("limit", limit);
  if (offset) backendParams.set("offset", offset);

  const queryString = backendParams.toString();
  const backendUrl = `${BACKEND_API}/occupation-groups${queryString ? `?${queryString}` : ""}`;

  console.log("[occupation-groups] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[occupation-groups] backend error body:", errorText);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[occupation-groups] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch occupation groups",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new occupation group
export async function POST(request: NextRequest) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const backendUrl = `${BACKEND_API}/occupation-groups`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[occupation-groups] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[occupation-groups] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create occupation group",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}