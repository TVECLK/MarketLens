import { describe, expect, it } from "vitest";
import {
  confirmsSourceName,
  findCaseOnlyNameMatch,
  findExactNameMatch,
  hasInnerWhitespace,
  hasOuterWhitespace,
  normalizeSourceName,
  visibleWhitespace,
} from "@/lib/source-name";

const sources = [{ name: "Ikman" }, { name: "TopJobs" }, { name: "Daily News" }];

describe("normalizeSourceName", () => {
  it("trims like the backend does", () => {
    expect(normalizeSourceName("  Ikman \t")).toBe("Ikman");
  });
});

describe("hasOuterWhitespace", () => {
  it("is true when saving would trim something", () => {
    expect(hasOuterWhitespace(" Ikman")).toBe(true);
    expect(hasOuterWhitespace("Ikman ")).toBe(true);
  });

  it("is false for a clean name or blank input", () => {
    expect(hasOuterWhitespace("Ikman")).toBe(false);
    expect(hasOuterWhitespace("   ")).toBe(false);
  });
});

describe("hasInnerWhitespace / visibleWhitespace", () => {
  it("flags and shows spaces inside the saved name only", () => {
    expect(hasInnerWhitespace("  New Crawler ")).toBe(true);
    expect(visibleWhitespace("  New Crawler ")).toBe("New·Crawler");
    expect(hasInnerWhitespace(" Ikman ")).toBe(false);
    expect(visibleWhitespace(" Ikman ")).toBe("Ikman");
  });
});

describe("findExactNameMatch", () => {
  it("matches the saved name exactly", () => {
    expect(findExactNameMatch(" Ikman ", sources)).toEqual({ name: "Ikman" });
    expect(findExactNameMatch("ikman", sources)).toBeUndefined();
  });
});

describe("findCaseOnlyNameMatch", () => {
  it("finds a name that differs only in letter case", () => {
    expect(findCaseOnlyNameMatch("ikman", sources)).toEqual({ name: "Ikman" });
    expect(findCaseOnlyNameMatch(" TOPJOBS ", sources)).toEqual({ name: "TopJobs" });
  });

  it("ignores an exact match, a different name and blank input", () => {
    expect(findCaseOnlyNameMatch("Ikman", sources)).toBeUndefined();
    expect(findCaseOnlyNameMatch("Rooster", sources)).toBeUndefined();
    expect(findCaseOnlyNameMatch("  ", sources)).toBeUndefined();
  });
});

describe("confirmsSourceName", () => {
  it("needs the exact, case-sensitive name", () => {
    expect(confirmsSourceName("Ikman", "Ikman")).toBe(true);
    expect(confirmsSourceName(" Ikman ", "Ikman")).toBe(true);
    expect(confirmsSourceName("ikman", "Ikman")).toBe(false);
    expect(confirmsSourceName("Ikma", "Ikman")).toBe(false);
  });
});
