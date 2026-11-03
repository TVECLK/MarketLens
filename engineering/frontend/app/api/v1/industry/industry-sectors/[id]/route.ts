import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from "@/lib/token";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// Get industry sector by id
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const backendUrl = `${BACKEND_API}/industry-sectors/${id}`;

  try {
    const response = await fetch(backendUrl);
    const data = await response.json();

    if (!response.ok) {
      console.error("[industry-sectors] backend error body:", data);
      return NextResponse.json(
        { error: "Backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-sectors] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch industry sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Update industry sector by id
export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const { id } = await params;
  const backendUrl = `${BACKEND_API}/industry-sectors/${id}`;

  try {
    const body = await request.json();

    const response = await fetch(backendUrl, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, },
      body: JSON.stringify(body),
    });

    const data = await response.json();

    if (!response.ok) {
      console.error("[industry-sectors] backend error body:", data);
      return NextResponse.json(
        { error: "Backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-sectors] update failed:", error);
    return NextResponse.json(
      {
        error: "Failed to update industry sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Delete industry sectors by id
export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const { id } = await params;
  const backendUrl = `${BACKEND_API}/industry-sectors/${id}`;

  try {
    const response = await fetch(backendUrl, { 
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}`, },
    });
    const data = await response.json();

    if (!response.ok) {
      console.error("[industry-sectors] backend error body:", data);
      return NextResponse.json(
        { error: "Backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[industry-sectors] delete failed:", error);
    return NextResponse.json(
      {
        error: "Failed to delete industry sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}