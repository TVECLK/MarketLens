import { beforeEach, describe, expect, it, vi } from "vitest";
import { NextRequest } from "next/server";

const sdk = { getSessionId: vi.fn(), getAccessToken: vi.fn() };
vi.mock("@thunderid/nextjs/server", () => ({ thunderid: vi.fn(async () => sdk) }));

import { withAuth } from "@/lib/with-auth";

const request = new NextRequest("http://localhost/api/v1/genders", { method: "POST" });

function signedIn(token: string) {
  sdk.getSessionId.mockResolvedValue("session-1");
  sdk.getAccessToken.mockResolvedValue(token);
}

describe("withAuth", () => {
  beforeEach(() => {
    sdk.getSessionId.mockReset();
    sdk.getAccessToken.mockReset();
  });

  it("responds 401 without calling the handler when there is no session", async () => {
    sdk.getSessionId.mockResolvedValue(undefined);
    const handler = vi.fn();

    const response = await withAuth(handler)(request, undefined);

    expect(response.status).toBe(401);
    expect(await response.json()).toEqual({ error: "Please sign in first" });
    expect(handler).not.toHaveBeenCalled();
    expect(sdk.getAccessToken).not.toHaveBeenCalled();
  });

  it("responds 401 when the SDK throws", async () => {
    sdk.getSessionId.mockRejectedValue(new Error("boom"));
    const handler = vi.fn();

    const response = await withAuth(handler)(request, undefined);

    expect(response.status).toBe(401);
    expect(handler).not.toHaveBeenCalled();
  });

  it("passes the token and the original route context to the handler", async () => {
    signedIn("abc123");
    const params = Promise.resolve({ id: "7" });
    const handler = vi.fn(async () => new Response("ok"));

    const response = await withAuth<{ params: Promise<{ id: string }> }>(handler)(request, { params });

    expect(await response.text()).toBe("ok");
    expect(sdk.getAccessToken).toHaveBeenCalledWith("session-1");
    expect(handler).toHaveBeenCalledWith(request, { params, token: "abc123" });
  });
});
