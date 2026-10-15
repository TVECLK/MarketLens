import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from "@/lib/token";

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// Get all vocational education levels
export async function GET() {
  const goUrl = `${GO_API}/vocational-educations`;

  console.log("[vocational-educations] fetching from Go:", goUrl);

  try {
    const response = await fetch(goUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[vocational-educations] Go error body:", errorText);
      return NextResponse.json(
        { error: "Go backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[vocational-educations] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch vocational educations",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new education level
export async function POST(request: NextRequest) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const goUrl = `${GO_API}/vocational-educations`;

  try {
    const body = await request.json();

    const response = await fetch(goUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    const data = await response.json();

    if (!response.ok) {
      console.error("[vocational-educations] Go error body:", data);
      return NextResponse.json(
        { error: "Go backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[vocational-educations] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create vocational education",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}