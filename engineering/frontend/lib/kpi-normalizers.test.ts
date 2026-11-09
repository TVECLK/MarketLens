import { describe, it, expect } from "vitest";
import { normTop, normChild } from "@/lib/kpi-normalizers";

describe("normTop", () => {
  it("maps id/name/code through and nulls out parent fields", () => {
    expect(normTop([{ id: 1, name: "Managers", code: "1" }])).toEqual([
      { id: 1, name: "Managers", code: "1", parentId: null, parentLabel: null },
    ]);
  });

  it("returns an empty array for undefined input", () => {
    expect(normTop(undefined)).toEqual([]);
  });

  it("returns an empty array for an empty list", () => {
    expect(normTop([])).toEqual([]);
  });
});

describe("normChild", () => {
  it("reads the parent id and builds a 'code — name' label when the parent is present", () => {
    const raw = [
      { id: 5, name: "Chief Executives", code: "11", major_group_id: 1, major_group: { name: "Managers", code: "1" } },
    ];
    expect(normChild(raw, "major_group_id", "major_group")).toEqual([
      { id: 5, name: "Chief Executives", code: "11", parentId: 1, parentLabel: "1 — Managers" },
    ]);
  });

  it("leaves the label null when the parent object isn't loaded", () => {
    const raw = [{ id: 5, name: "Chief Executives", code: "11", major_group_id: 1, major_group: null }];
    expect(normChild(raw, "major_group_id", "major_group")[0].parentLabel).toBeNull();
  });

  it("returns an empty array for undefined input", () => {
    expect(normChild(undefined, "major_group_id", "major_group")).toEqual([]);
  });
});
