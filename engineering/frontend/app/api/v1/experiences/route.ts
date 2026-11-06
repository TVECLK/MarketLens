import { NextResponse } from "next/server";
import { withAuth } from '@/lib/with-auth'

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get all experiences
export async function GET() {
  const backendUrl = `${BACKEND_API}/experiences`;

  console.log("[experiences] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[experiences] backend error body:", errorText);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorText },
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
export const POST = withAuth(async (request, { token }) => {
  const backendUrl = `${BACKEND_API}/experiences`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[experiences] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

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
});