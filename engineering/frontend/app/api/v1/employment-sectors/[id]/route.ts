import { NextRequest, NextResponse } from "next/server";
import { getAccessToken } from '@/lib/token';

const GO_API = `${process.env.GO_BACKEND_URL}/api/v1`;

// Get employment sector by id
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const goUrl = `${GO_API}/employment-sectors/${id}`;

  try {
    const response = await fetch(goUrl);
    const data = await response.json();

    if (!response.ok) {
      console.error("[employment-sectors] Go error body:", data);
      return NextResponse.json(
        { error: "Go backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[employment-sectors] fetch failed:", error);
    return NextResponse.json(
      {
        error: "Failed to fetch employment sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Update employment sector by id
export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const { id } = await params;
  const goUrl = `${GO_API}/employment-sectors/${id}`;

  try {
    const body = await request.json();

    const response = await fetch(goUrl, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`,  },
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

    return NextResponse.json(data);
  } catch (error) {
    console.error("[employment-sectors] update failed:", error);
    return NextResponse.json(
      {
        error: "Failed to update employment sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}

// Delete employment sector by id
export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const token = await getAccessToken()
  if (!token) {
    return NextResponse.json({ error: 'Please sign in first' }, { status: 401 })
  }

  const { id } = await params;
  const goUrl = `${GO_API}/employment-sectors/${id}`;

  try {
    const response = await fetch(goUrl, { 
      method: "DELETE", 
      headers: { Authorization: `Bearer ${token}` }
    });
    const data = await response.json();

    if (!response.ok) {
      console.error("[employment-sectors] Go error body:", data);
      return NextResponse.json(
        { error: "Go backend returned an error", details: data },
        { status: response.status }
      );
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error("[employment-sectors] delete failed:", error);
    return NextResponse.json(
      {
        error: "Failed to delete employment sector",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 500 }
    );
  }
}