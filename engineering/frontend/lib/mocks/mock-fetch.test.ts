import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mockFetch } from "@/lib/mocks/mock-fetch";
import { resetMockSources } from "@/lib/mocks/source-data";
import { getApiErrorMessage } from "@/lib/errors";
import type { Source, SourceListResponse } from "@/types/source";

const URL = "/api/v1/sources";

/** Calls the mock, skipping its simulated latency. */
async function call(path: string, init?: RequestInit) {
  const pending = mockFetch(path, init);
  await vi.advanceTimersByTimeAsync(1000);
  const res = await pending;
  return { status: res.status, body: await res.json() };
}

function send(method: string, body: unknown): RequestInit {
  return { method, headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) };
}

beforeEach(() => {
  vi.useFakeTimers();
  resetMockSources();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("GET /sources", () => {
  it("lists live sources only, as { count, sources }", async () => {
    const { status, body } = await call(URL);
    const list = body as SourceListResponse;
    expect(status).toBe(200);
    expect(list.count).toBe(list.sources.length);
    expect(list.sources.map((s) => s.name)).not.toContain("Sunday Observer"); // deleted
  });

  it("applies the type and is-active filters", async () => {
    const auto = (await call(`${URL}?type=auto`)).body as SourceListResponse;
    expect(auto.sources.every((s) => s.type === "auto")).toBe(true);
    const inactive = (await call(`${URL}?is-active=false`)).body as SourceListResponse;
    expect(inactive.sources.every((s) => !s.is_active)).toBe(true);
  });

  it("rejects an unknown type with a 400", async () => {
    const { status, body } = await call(`${URL}?type=rss`);
    expect(status).toBe(400);
    expect(getApiErrorMessage(body, "")).toContain("invalid type");
  });
});

describe("POST /sources", () => {
  it("creates a trimmed, active source by default", async () => {
    const { status, body } = await call(URL, send("POST", { name: "  The Island ", type: "manual" }));
    expect(status).toBe(201);
    expect(body).toMatchObject({ name: "The Island", type: "manual", is_active: true });
  });

  it("keeps is_active false when asked", async () => {
    const { body } = await call(URL, send("POST", { name: "NewCrawler", type: "auto", is_active: false }));
    expect((body as Source).is_active).toBe(false);
  });

  it("rejects a live name as 'already exists'", async () => {
    const { status, body } = await call(URL, send("POST", { name: "Ikman", type: "auto" }));
    expect(status).toBe(400);
    expect(getApiErrorMessage(body, "")).toContain("already exists");
  });

  it("restores a deleted name with the same id and the new settings", async () => {
    const { status, body } = await call(URL, send("POST", { name: "Sunday Observer", type: "manual", is_active: false }));
    expect(status).toBe(201);
    expect(body).toMatchObject({ id: 8, name: "Sunday Observer", is_active: false });
  });
});

describe("PUT /sources/:id", () => {
  it("renames a manual source", async () => {
    const { status, body } = await call(`${URL}/6`, send("PUT", { name: "Daily News LK" }));
    expect(status).toBe(200);
    expect((body as Source).name).toBe("Daily News LK");
  });

  it("refuses to rename an auto source", async () => {
    const { status, body } = await call(`${URL}/1`, send("PUT", { name: "Ikman.lk" }));
    expect(status).toBe(400);
    expect(getApiErrorMessage(body, "")).toContain("only manual sources can be renamed");
  });

  it("refuses a name another source has, deleted ones included", async () => {
    expect((await call(`${URL}/6`, send("PUT", { name: "Lankadeepa" }))).status).toBe(400);
    expect((await call(`${URL}/6`, send("PUT", { name: "Sunday Observer" }))).status).toBe(400);
  });
});

describe("PATCH /sources/:id/status", () => {
  it("enables or disables a source", async () => {
    const { body } = await call(`${URL}/1/status`, send("PATCH", { is_active: false }));
    expect((body as Source).is_active).toBe(false);
  });

  it("is a 404 for a deleted source", async () => {
    expect((await call(`${URL}/8/status`, send("PATCH", { is_active: true }))).status).toBe(404);
  });
});

describe("DELETE /sources/:id", () => {
  it("soft-deletes, then 404s on a second delete", async () => {
    expect((await call(`${URL}/1`, { method: "DELETE" })).status).toBe(200);
    const list = (await call(URL)).body as SourceListResponse;
    expect(list.sources.map((s) => s.name)).not.toContain("Ikman");
    expect((await call(`${URL}/1`, { method: "DELETE" })).status).toBe(404);
  });
});

describe("unknown routes", () => {
  it("answers 404 instead of throwing", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    expect((await call("/api/v1/nope")).status).toBe(404);
  });
});
