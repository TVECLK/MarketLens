import { describe, expect, it, vi } from "vitest";
import { fetchJson } from "./fetch-json";

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status });

describe("fetchJson", () => {
  it("returns the parsed JSON on success", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(json({ n: 1 }));
    await expect(fetchJson("/a")).resolves.toEqual({ n: 1 });
  });

  it("passes the signal to fetch", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(json({}));
    const signal = new AbortController().signal;
    await fetchJson("/a", { signal });
    expect(fetchSpy).toHaveBeenCalledWith("/a", { signal });
  });

  it("throws the message from the error body", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(json({ error: "missing required scope: genders:create" }, 403));
    await expect(fetchJson("/a")).rejects.toThrow("missing required scope: genders:create");
  });

  it("uses the given errorMessage when the error body has no message", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(json({}, 500));
    await expect(fetchJson("/a", { errorMessage: "Failed to fetch a" })).rejects.toThrow("Failed to fetch a");
  });

  it("falls back to 'Request failed: <url>' when the error body isn't JSON", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(new Response("oops", { status: 502 }));
    await expect(fetchJson("/a")).rejects.toThrow("Request failed: /a");
  });
});
