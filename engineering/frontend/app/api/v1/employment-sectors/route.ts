import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from '@/lib/token';

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// Get all employment sectors
export async function GET() {
  const goUrl = `${GO_API}/employment-sectors`;

  console.log("[employment-sectors] fetching from Go:", goUrl);

  try {
    const response = await fetch(goUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[employment-sectors] Go error body:", errorText);
      return NextResponse.json(
        { error: "Go backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[employment-sectors] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch employment sectors",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new employment sector
export async function POST(request: NextRequest) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const goUrl = `${GO_API}/employment-sectors`;

  try {
    const body = await request.json();

    const response = await fetch(goUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    const data = await response.json();

    if (!response.ok) {
      console.error("[employment-sectors] Go error body:", data);
      return NextResponse.json(
        { error: "Go backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[employment-sectors] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create employment sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}