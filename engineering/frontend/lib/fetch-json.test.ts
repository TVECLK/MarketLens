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

  it("sends the method and a JSON body", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(json({ id: 1 }, 201));
    await expect(fetchJson("/a", { method: "POST", body: { name: "x" } })).resolves.toEqual({ id: 1 });
    expect(fetchSpy).toHaveBeenCalledWith("/a", {
      signal: undefined,
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: "x" }),
    });
  });

  it("sends a method without a body or Content-Type when there's no body", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(json({ message: "ok" }));
    await fetchJson("/a/1", { method: "DELETE" });
    expect(fetchSpy).toHaveBeenCalledWith("/a/1", { signal: undefined, method: "DELETE" });
  });

  it("uses the given fetcher instead of fetch", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(json({ n: 2 }));
    await expect(fetchJson("/a", { fetcher })).resolves.toEqual({ n: 2 });
    expect(fetcher).toHaveBeenCalledWith("/a", { signal: undefined });
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it("throws the fetcher's error message the same way", async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValue(json({ details: { error: "already exists" } }, 400));
    await expect(fetchJson("/a", { method: "POST", body: {}, fetcher })).rejects.toThrow("already exists");
  });
});
