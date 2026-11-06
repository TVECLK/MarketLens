import { NextResponse } from "next/server";

const BACKEND_API = process.env.BACKEND_API_URL;

export async function GET() {
  console.log("[readyz] checking backend health:", `${BACKEND_API}/healthz`);

  try {
    const response = await fetch(`${BACKEND_API}/healthz`, {
      signal: AbortSignal.timeout(2000),
    });

    if (!response.ok) {
      console.error("[readyz] backend returned non-OK status:", response.status);
      return NextResponse.json(
        { status: "not ready", reason: "backend unhealthy" },
        { status: 503 }
      );
    }

    return NextResponse.json({ status: "ready" });
  } catch (error) {
    console.error("[readyz] backend unreachable:", error);
    return NextResponse.json(
      {
        status: "not ready",
        reason: "backend unreachable",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 503 }
    );
  }
}