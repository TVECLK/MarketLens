import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from "@/lib/token";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// Get all sub major groups
export async function GET() {
  const goUrl = `${GO_API}/sub-major-groups`;

  console.log("[sub-major-groups] fetching from Go:", goUrl);

  try {
    const response = await fetch(goUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[sub-major-groups] Go error body:", errorText);
      return NextResponse.json(
        { error: "Go backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[sub-major-groups] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch sub major groups",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new sub major group
export async function POST(request: NextRequest) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const goUrl = `${GO_API}/sub-major-groups`;

  try {
    const body = await request.json();

    const response = await fetch(goUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    const data = await response.json();

    if (!response.ok) {
      console.error("[sub-major-groups] Go error body:", data);
      return NextResponse.json(
        { error: "Go backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[sub-major-groups] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create sub major group",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}