import { NextRequest, NextResponse } from "next/server";
import { thunderid } from "@thunderid/nextjs/server";

/** The signed-in user's access token, or undefined when there is no session. */
export async function getAccessToken(): Promise<string | undefined> {
  try {
    const sdk = await thunderid();
    const sessionId = await sdk.getSessionId();
    if (!sessionId) return undefined;
    return await sdk.getAccessToken(sessionId);
  } catch {
    return undefined;
  }
}

/**
 * Wraps a route handler so it only runs for a signed-in user.
 *
 * Responds 401 when there is no access token; otherwise calls the handler with
 * the token added to the route context. This only checks that a session exists:
 * the backend must still validate the token and the user's permissions.
 */
export function withAuth<Ctx = unknown>(
  handler: (request: NextRequest, ctx: Ctx & { token: string }) => Promise<Response>
) {
  return async (request: NextRequest, ctx: Ctx): Promise<Response> => {
    const token = await getAccessToken();
    if (!token) {
      return NextResponse.json({ error: "Please sign in first" }, { status: 401 });
    }
    return handler(request, { ...ctx, token });
  };
}
