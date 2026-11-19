import { NextResponse } from "next/server";

const BACKEND_API = `${process.env.BACKEND_API_URL}/api/v1`;

// GET /api/v1/dashboard-card-data
// Returns all-time headline numbers: source count, total vacancy count, industry count.
export async function GET() {
  const res = await fetch(`${BACKEND_API}/dashboard-card-data`, { cache: "no-store" });
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    return NextResponse.json(data ?? { error: "Failed to fetch dashboard card data" }, { status: res.status });
  }
  return NextResponse.json(data);
}

