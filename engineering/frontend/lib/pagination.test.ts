import { describe, expect, it } from "vitest";
import { clampPage } from "./pagination";

describe("clampPage", () => {
  it("leaves the page alone while the total is unknown", () => {
    expect(clampPage(3, undefined, 20)).toBe(3);
  });

  it("keeps a page that exists", () => {
    expect(clampPage(2, 40, 20)).toBe(2);
  });

  it("moves back to the last page when the page is past the end", () => {
    // 41 rows -> 3 pages; deleting the only row on page 3 leaves 40 rows.
    expect(clampPage(3, 40, 20)).toBe(2);
  });

  it("falls back to page 1 for an empty list", () => {
    expect(clampPage(2, 0, 20)).toBe(1);
  });

  it("never goes below page 1", () => {
    expect(clampPage(0, 100, 20)).toBe(1);
  });
});
