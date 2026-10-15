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
      {
        id: 5,
        name: "Chief Executives",
        code: "11",
        major_group_id: 1,
        major_group: { code: "1", name: "Managers" },
      },
    ];

    expect(normChild(raw, "major_group_id", "major_group")).toEqual([
      {
        id: 5,
        name: "Chief Executives",
        code: "11",
        parentId: 1,
        parentLabel: "1 — Managers",
      },
    ]);
  });

  it("sets parentLabel to null when the embedded parent is null", () => {
    const raw = [
      {
        id: 5,
        name: "Chief Executives",
        code: "11",
        major_group_id: 1,
        major_group: null,
      },
    ];

    expect(normChild(raw, "major_group_id", "major_group")[0].parentLabel).toBeNull();
  });

  it("works with a different parent field name (e.g. the industry hierarchy)", () => {
    const raw = [
      {
        id: 9,
        name: "Crop Production",
        code: "011",
        industry_sector_id: 2,
        industry_sector: { code: "A", name: "Agriculture" },
      },
    ];

    const result = normChild(raw, "industry_sector_id", "industry_sector");
    expect(result[0].parentId).toBe(2);
    expect(result[0].parentLabel).toBe("A — Agriculture");
  });

  it("returns an empty array for undefined input", () => {
    expect(normChild(undefined, "major_group_id", "major_group")).toEqual([]);
  });
});
