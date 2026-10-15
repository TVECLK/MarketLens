import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from '@/lib/token'

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// Get all experiences
export async function GET() {
  const goUrl = `${GO_API}/experiences`;

  console.log("[experiences] fetching from Go:", goUrl);

  try {
    const response = await fetch(goUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[experiences] Go error body:", errorText);
      return NextResponse.json(
        { error: "Go backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[experiences] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch experiences",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new experience level
export async function POST(request: NextRequest) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const goUrl = `${GO_API}/experiences`;

  try {
    const body = await request.json();

    const response = await fetch(goUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    const data = await response.json();

    if (!response.ok) {
      console.error("[experiences] Go error body:", data);
      return NextResponse.json(
        { error: "Go backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[experiences] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create experience",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}