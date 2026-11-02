import { describe, expect, it, vi } from "vitest";
import { fetchSections } from "./bff-aggregate";

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status });

// Builds a fake fetch that answers by URL.
function fakeFetch(routes: Record<string, () => Response | Promise<Response>>) {
  return vi.fn(async (url: string | URL | Request) => {
    const handler = routes[String(url)];
    if (!handler) throw new Error(`unexpected url ${String(url)}`);
    return handler();
  }) as unknown as typeof fetch;
}

const requests = [
  { key: "a", url: "/a" },
  { key: "b", url: "/b" },
  { key: "c", url: "/c" },
];

describe("fetchSections", () => {
  it("returns every section's data when all succeed", async () => {
    const result = await fetchSections(
      requests,
      fakeFetch({ "/a": () => json({ n: 1 }), "/b": () => json({ n: 2 }), "/c": () => json({ n: 3 }) }),
    );
    expect(result.data).toEqual({ a: { n: 1 }, b: { n: 2 }, c: { n: 3 } });
    expect(result.errors).toEqual({});
    expect(result.allFailed).toBe(false);
    expect(result.firstError).toBeNull();
  });

  it("keeps the successful sections when one fails", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    const result = await fetchSections(
      requests,
      fakeFetch({
        "/a": () => json({ n: 1 }),
        "/b": () => json({ error: "boom" }, 500),
        "/c": () => json({ n: 3 }),
      }),
    );
    expect(result.data).toEqual({ a: { n: 1 }, b: null, c: { n: 3 } });
    expect(result.errors).toEqual({ b: { status: 500, message: "boom" } });
    expect(result.allFailed).toBe(false);
  });

  it("reports allFailed and forwards the first error when every section fails", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    const result = await fetchSections(
      requests,
      fakeFetch({
        "/a": () => json({ error: "bad dates" }, 400),
        "/b": () => json({ error: "other" }, 500),
        "/c": () => json({ error: "other" }, 500),
      }),
    );
    expect(result.allFailed).toBe(true);
    expect(result.firstError).toEqual({ status: 400, body: { error: "bad dates" } });
  });

  it("treats a network failure as a 502 for that section", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    const result = await fetchSections(
      requests,
      fakeFetch({
        "/a": () => json({ n: 1 }),
        "/b": () => Promise.reject(new Error("ECONNREFUSED")),
        "/c": () => json({ n: 3 }),
      }),
    );
    expect(result.errors.b).toEqual({ status: 502, message: "ECONNREFUSED" });
    expect(result.data.b).toBeNull();
  });

  it("falls back to a generic message when the error body isn't JSON", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    const result = await fetchSections(
      [{ key: "a", url: "/a" }],
      fakeFetch({ "/a": () => new Response("oops", { status: 503 }) }),
    );
    expect(result.errors.a).toEqual({ status: 503, message: "Go backend returned an error from a" });
  });

  it("treats an unparseable success body as a failed section", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    const result = await fetchSections(
      [{ key: "a", url: "/a" }],
      fakeFetch({ "/a": () => new Response("not json", { status: 200 }) }),
    );
    expect(result.errors.a?.status).toBe(502);
    expect(result.data.a).toBeNull();
  });

  it("ignores an optional section's failure: null data, no error, not counted", async () => {
    const result = await fetchSections(
      [
        { key: "a", url: "/a" },
        { key: "children", url: "/children", optional: true },
      ],
      fakeFetch({ "/a": () => json({ n: 1 }), "/children": () => json({ error: "leaf" }, 404) }),
    );
    expect(result.data).toEqual({ a: { n: 1 }, children: null });
    expect(result.errors).toEqual({});
    expect(result.allFailed).toBe(false);
  });

  it("is allFailed when every required section fails even if an optional one succeeds", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    const result = await fetchSections(
      [
        { key: "a", url: "/a" },
        { key: "children", url: "/children", optional: true },
      ],
      fakeFetch({ "/a": () => json({ error: "x" }, 500), "/children": () => json({ n: 9 }) }),
    );
    expect(result.allFailed).toBe(true);
    expect(result.data.children).toEqual({ n: 9 });
  });
});
