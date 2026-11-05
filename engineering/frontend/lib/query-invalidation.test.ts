import { describe, expect, it } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { OCCUPATION_QUERY_KEYS, INDUSTRY_QUERY_KEYS } from "@/lib/kpi-config";
import { invalidateAfterKpiChange, relatedListKeys } from "./query-invalidation";

describe("relatedListKeys", () => {
  it("returns every occupation level for an occupation level", () => {
    expect(relatedListKeys("minor-groups")).toEqual(OCCUPATION_QUERY_KEYS);
  });

  it("returns every industry level for an industry level", () => {
    expect(relatedListKeys("industry-classes")).toEqual(INDUSTRY_QUERY_KEYS);
  });

  it("returns only its own key for a flat KPI", () => {
    expect(relatedListKeys("genders")).toEqual(["genders"]);
  });
});

describe("invalidateAfterKpiChange", () => {
  const seeded = [
    ["genders", "list", {}],
    ["major-groups", "list", {}],
    ["occupation", "major-groups", "2026-01-01", "2026-06-01"],
    ["industry", "industry-sectors", "2026-01-01", "2026-06-01"],
    ["analysis", "occupation", "major-group", 1, "2026-01-01", "2026-06-01"],
    ["dashboard", "overview", "2026-01-01", "2026-06-01"],
    ["occupation-skills", "major-group", 1, "2026-01-01", "2026-06-01"],
  ];

  function makeClient() {
    const client = new QueryClient();
    for (const key of seeded) client.setQueryData(key, { ok: true });
    return client;
  }

  const isInvalidated = (client: QueryClient, key: unknown[]) =>
    client.getQueryState(key)?.isInvalidated ?? false;

  it("invalidates the dashboard and analysis queries as well as the admin list", () => {
    const client = makeClient();
    invalidateAfterKpiChange(client, ["genders"]);

    expect(isInvalidated(client, ["genders", "list", {}])).toBe(true);
    for (const key of seeded.slice(2)) expect(isInvalidated(client, key)).toBe(true);
  });

  it("leaves unrelated admin lists alone", () => {
    const client = makeClient();
    invalidateAfterKpiChange(client, ["genders"]);

    expect(isInvalidated(client, ["major-groups", "list", {}])).toBe(false);
  });
});
