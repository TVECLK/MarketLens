import { describe, it, expect } from "vitest";
import { normTop } from "@/lib/kpi-normalizers";

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
