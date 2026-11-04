import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from '@/lib/token';

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get all employment sectors
export async function GET() {
  const backendUrl = `${BACKEND_API}/employment-sectors`;

  console.log("[employment-sectors] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[employment-sectors] backend error body:", errorText);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorText },
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

  const backendUrl = `${BACKEND_API}/employment-sectors`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[employment-sectors] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

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