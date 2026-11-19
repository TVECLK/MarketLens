import { NextResponse } from "next/server";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// GET /api/v1/dashboard-card-data
// Returns all-time headline numbers: source count, total vacancy count, industry count.
export async function GET() {
  const res = await fetch(`${BACKEND_API}/dashboard-card-data`, { cache: "no-store" });
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: "Failed to fetch dashboard card data" }));
    return NextResponse.json(error, { status: res.status });
  }
  const data = await res.json();
  return NextResponse.json(data);
}

