import { describe, expect, it } from "vitest";
import { selectedIdAt } from "./use-hierarchy-drilldown";

const LEVELS = ["sector", "division", "group"] as const;

describe("selectedIdAt", () => {
  it("returns the id selected at the given level", () => {
    expect(selectedIdAt([2, 11, 40], LEVELS, "sector")).toBe(2);
    expect(selectedIdAt([2, 11, 40], LEVELS, "group")).toBe(40);
  });

  it("returns null for a level that isn't selected yet", () => {
    expect(selectedIdAt([2], LEVELS, "division")).toBeNull();
    expect(selectedIdAt([], LEVELS, "sector")).toBeNull();
  });

  it("returns null for a level that isn't in the list", () => {
    expect(selectedIdAt([2, 11], LEVELS, "class" as never)).toBeNull();
  });

  it("follows the level by name when the list is reordered", () => {
    const reordered = ["group", "sector", "division"] as const;
    expect(selectedIdAt([40, 2, 11], reordered, "sector")).toBe(2);
  });
});
