import { describe, expect, it } from "vitest";
import {
  isDateString,
  isUsableDateParam,
  parseEnumParam,
  parseHierarchyIds,
  serializeHierarchyIds,
} from "./url-params";

const LEVELS = ["major-group", "sub-major-group", "minor-group"] as const;
const reader = (query: string) => new URLSearchParams(query);

describe("isDateString", () => {
  it("accepts YYYY-MM-DD strings, ignoring surrounding whitespace", () => {
    expect(isDateString("2026-01-31")).toBe(true);
    expect(isDateString(" 2026-01-31 ")).toBe(true);
  });

  it("rejects non-strings", () => {
    expect(isDateString(null)).toBe(false);
    expect(isDateString(undefined)).toBe(false);
    expect(isDateString(20260131)).toBe(false);
    expect(isDateString(new Date("2026-01-31"))).toBe(false);
  });

  it("rejects other formats, blanks and padded wrong formats", () => {
    expect(isDateString("31-01-2026")).toBe(false);
    expect(isDateString("2026-1-31")).toBe(false);
    expect(isDateString("2026/01/31")).toBe(false);
    expect(isDateString("")).toBe(false);
    expect(isDateString("      ")).toBe(false);
    expect(isDateString(" 31-01-2026 ")).toBe(false);
  });

  it("rejects dates that don't exist on the calendar", () => {
    expect(isDateString("2026-02-30")).toBe(false);
    expect(isDateString("2026-04-31")).toBe(false);
    expect(isDateString("2026-02-29")).toBe(false);
    expect(isDateString("2026-13-01")).toBe(false);
    expect(isDateString("2026-00-10")).toBe(false);
    expect(isDateString("2026-01-00")).toBe(false);
  });

  it("accepts real edge-of-month and leap days", () => {
    expect(isDateString("2024-02-29")).toBe(true);
    expect(isDateString("2026-12-31")).toBe(true);
    expect(isDateString("2026-04-30")).toBe(true);
  });
});

describe("isUsableDateParam", () => {
  it("accepts ISO dates and the empty string, rejects the rest", () => {
    expect(isUsableDateParam("2026-01-31")).toBe(true);
    expect(isUsableDateParam("")).toBe(true);
    expect(isUsableDateParam(null)).toBe(false);
    expect(isUsableDateParam("31-01-2026")).toBe(false);
  });

  it("ignores surrounding whitespace", () => {
    expect(isUsableDateParam("      ")).toBe(true);
    expect(isUsableDateParam(" 2026-01-31 ")).toBe(true);
    expect(isUsableDateParam(" 31-01-2026 ")).toBe(false);
  });

  it("rejects dates that don't exist on the calendar", () => {
    expect(isUsableDateParam("2026-02-30")).toBe(false);
    expect(isUsableDateParam(" 2026-04-31 ")).toBe(false);
    expect(isUsableDateParam("2024-02-29")).toBe(true);
  });
});

describe("parseEnumParam", () => {
  it("returns allowed values and falls back otherwise", () => {
    expect(parseEnumParam("table", ["chart", "table"], "chart")).toBe("table");
    expect(parseEnumParam("grid", ["chart", "table"], "chart")).toBe("chart");
    expect(parseEnumParam(null, ["chart", "table"], "chart")).toBe("chart");
    expect(parseEnumParam("", ["chart", "table"], "chart")).toBe("chart");
    expect(parseEnumParam("   ", ["chart", "table"], "chart")).toBe("chart");
    expect(parseEnumParam("table", [], "chart")).toBe("chart");
    expect(parseEnumParam("", [], "chart")).toBe("chart");
    expect(parseEnumParam("other", ["chart", "table"], "")).toBe("");
  });
});

describe("parseHierarchyIds", () => {
  it("reads one id per level in order", () => {
    expect(parseHierarchyIds(reader("major-group=2&sub-major-group=11"), LEVELS)).toEqual([2, 11]);
    expect(
      parseHierarchyIds(reader("major-group=2&sub-major-group=11&minor-group=12"), LEVELS),
    ).toEqual([2, 11, 12]);
  });

  it("stops at the first missing level", () => {
    expect(parseHierarchyIds(reader("major-group=2&minor-group=43"), LEVELS)).toEqual([2]);
  });

  it("stops at a non-numeric id", () => {
    expect(parseHierarchyIds(reader("major-group=abc&sub-major-group=1"), LEVELS)).toEqual([]);
    expect(parseHierarchyIds(reader("major-group=1&sub-major-group=abc"), LEVELS)).toEqual([1]);
    expect(parseHierarchyIds(reader("major-group=abc&sub-major-group=abc"), LEVELS)).toEqual([]);
    expect(parseHierarchyIds(reader("major-group=-1"), LEVELS)).toEqual([]);
  });
});

describe("serializeHierarchyIds", () => {
  it("sets selected levels and clears deeper ones", () => {
    expect(serializeHierarchyIds([2, 11], LEVELS)).toEqual({
      "major-group": "2",
      "sub-major-group": "11",
      "minor-group": null,
    });
    expect(serializeHierarchyIds([2], LEVELS)).toEqual({
      "major-group": "2",
      "sub-major-group": null,
      "minor-group": null,
    });
    expect(serializeHierarchyIds([], LEVELS)).toEqual({
      "major-group": null,
      "sub-major-group": null,
      "minor-group": null,
    });
  });

  it("round-trips through parseHierarchyIds", () => {
    const params = new URLSearchParams();
    for (const [k, v] of Object.entries(serializeHierarchyIds([4, 5, 6], LEVELS))) {
      if (v !== null) params.set(k, v);
    }
    expect(parseHierarchyIds(params, LEVELS)).toEqual([4, 5, 6]);
  });
});
