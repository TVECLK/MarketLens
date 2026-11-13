import { NextResponse } from "next/server";
import { withAuth } from "@/lib/with-auth";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Enable or disable a source
export const PATCH = withAuth<{ params: Promise<{ id: string }> }>(async (request, { params, token }) => {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/sources/${id}/status`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "PATCH",
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
    console.error("[sources] status update failed:", error);
    return NextResponse.json(
      { error: "Failed to update source status" },
      { status: 500 }
    );
  }
});
