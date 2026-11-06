import { NextRequest, NextResponse } from "next/server";
import { withAuth } from '@/lib/with-auth'

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get education level by id
export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/education-levels/${id}`;

  try {
    const response = await fetch(backendUrl);
    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[education-levels] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[education-levels] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch education level",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Update education level by id
export const PUT = withAuth<{ params: Promise<{ id: string }> }>(async (request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/education-levels/${id}`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[education-levels] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[education-levels] update failed:", error);
    return NextResponse.json(
      {
        error: "Failed to update education level",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});

// Delete education level by id
export const DELETE = withAuth<{ params: Promise<{ id: string }> }>(async (_request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/education-levels/${id}`;

  try {
    const response = await fetch(backendUrl, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[education-levels] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();

    return NextResponse.json(data);
  } catch (error) {
    console.error("[education-levels] delete failed:", error);
    return NextResponse.json(
      {
        error: "Failed to delete education level",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});