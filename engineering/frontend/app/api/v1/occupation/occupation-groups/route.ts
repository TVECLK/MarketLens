import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from "@/lib/token";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// Get all occupation groups with pagination
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const limit = searchParams.get("limit");
  const offset = searchParams.get("offset");

  const goParams = new URLSearchParams();
  if (limit) goParams.set("limit", limit);
  if (offset) goParams.set("offset", offset);

  const queryString = goParams.toString();
  const goUrl = `${GO_API}/occupation-groups${queryString ? `?${queryString}` : ""}`;

  console.log("[occupation-groups] fetching from Go:", goUrl);

  try {
    const response = await fetch(goUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[occupation-groups] Go error body:", errorText);
      return NextResponse.json(
        { error: "Go backend returned an error", details: errorText },
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

  const goUrl = `${GO_API}/occupation-groups`;

  try {
    const body = await request.json();

    const response = await fetch(goUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    const data = await response.json();

    if (!response.ok) {
      console.error("[occupation-groups] Go error body:", data);
      return NextResponse.json(
        { error: "Go backend returned an error", details: data },
        { status: response.status }
      );
    }

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