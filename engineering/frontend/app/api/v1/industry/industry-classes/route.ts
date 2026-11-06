import { NextResponse } from "next/server";
import { withAuth } from "@/lib/with-auth";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get all industry classes
export async function GET() {
  const backendUrl = `${BACKEND_API}/industry-classes`;

  console.log("[industry-classes] fetching from backend:", backendUrl);

  try {
    const response = await fetch(backendUrl);

    if (!response.ok) {
      const errorText = await response.text();
      console.error("[industry-classes] backend error body:", errorText);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-classes] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch industry classes",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Create new industry class
export const POST = withAuth(async (request, { token }) => {
  const backendUrl = `${BACKEND_API}/industry-classes`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[industry-classes] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data, { status: 201 });
  } catch (error) {
    console.error("[industry-classes] create failed:", error);
    return NextResponse.json(
      {
        error: "Failed to create industry class",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});