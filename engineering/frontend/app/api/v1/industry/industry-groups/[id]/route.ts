import { NextRequest, NextResponse } from "next/server";
import { withAuth } from "@/lib/with-auth";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get industry groups by id
export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/industry-groups/${id}`;

  try {
    const response = await fetch(backendUrl);
    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[industry-groups] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-groups] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch industry group",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Update industry groups by id
export const PUT = withAuth<{ params: Promise<{ id: string }> }>(async (request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/industry-groups/${id}`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[industry-groups] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-groups] update failed:", error);
    return NextResponse.json(
      {
        error: "Failed to update industry group",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});

// Delete industry groups by id
export const DELETE = withAuth<{ params: Promise<{ id: string }> }>(async (_request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/industry-groups/${id}`;

  try {
    const response = await fetch(backendUrl, { 
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}`, }, 
    });
    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[industry-groups] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-groups] delete failed:", error);
    return NextResponse.json(
      {
        error: "Failed to delete industry group",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});