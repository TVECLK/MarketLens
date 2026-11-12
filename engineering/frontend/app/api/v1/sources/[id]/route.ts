import { NextResponse } from "next/server";
import { withAuth } from "@/lib/with-auth";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

type Ctx = { params: Promise<{ id: string }> };

// Get source by id (needs `sources:read`, so the token is sent)
export const GET = withAuth<Ctx>(async (_request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/sources/${id}`;

  try {
    const response = await fetch(backendUrl, {
      headers: { Authorization: `Bearer ${token}` },
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[sources] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[sources] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch source",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});

// Rename a source (manual sources only; the backend rejects auto)
export const PUT = withAuth<Ctx>(async (request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/sources/${id}`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[sources] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[sources] rename failed:", error);
    return NextResponse.json(
      {
        error: "Failed to rename source",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});

// Soft-delete a source
export const DELETE = withAuth<Ctx>(async (_request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/sources/${id}`;

  try {
    const response = await fetch(backendUrl, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });

    if (!response.ok) {
      const errorBody = await response.json().catch(() => null);
      console.error("[sources] backend error body:", errorBody);
      return NextResponse.json(
        { error: "Backend returned an error", details: errorBody },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("[sources] delete failed:", error);
    return NextResponse.json(
      {
        error: "Failed to delete source",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
});
