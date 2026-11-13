import { describe, expect, it } from "vitest";
import {
  confirmsSourceName,
  displaySourceName,
  findExactNameMatch,
  hasInnerWhitespace,
  hasOuterWhitespace,
  hasUppercase,
  normalizeSourceName,
  visibleWhitespace,
} from "@/lib/source-name";

const sources = [{ name: "ikman" }, { name: "TopJobs" }, { name: "daily news" }];

describe("normalizeSourceName", () => {
  it("trims and lowercases", () => {
    expect(normalizeSourceName("  Ikman \t")).toBe("ikman");
    expect(normalizeSourceName("JobSiteLK")).toBe("jobsitelk");
  });
});

describe("displaySourceName", () => {
  it("capitalizes the first letter", () => {
    expect(displaySourceName("ikman")).toBe("Ikman");
    expect(displaySourceName("daily news")).toBe("Daily news");
  });

  it("leaves the rest and empty names alone", () => {
    expect(displaySourceName("TopJobs")).toBe("TopJobs");
    expect(displaySourceName("")).toBe("");
  });
});

describe("hasOuterWhitespace", () => {
  it("is true when saving would trim something", () => {
    expect(hasOuterWhitespace(" ikman")).toBe(true);
    expect(hasOuterWhitespace("ikman ")).toBe(true);
  });

  it("is false for a clean name, whatever its case, or blank input", () => {
    expect(hasOuterWhitespace("Ikman")).toBe(false);
    expect(hasOuterWhitespace("   ")).toBe(false);
  });
});

describe("hasUppercase", () => {
  it("is true when saving would lowercase something", () => {
    expect(hasUppercase(" Ikman ")).toBe(true);
  });

  it("is false for lowercase or blank input", () => {
    expect(hasUppercase(" ikman ")).toBe(false);
    expect(hasUppercase("")).toBe(false);
  });
});

describe("hasInnerWhitespace / visibleWhitespace", () => {
  it("flags and shows spaces inside the saved name only", () => {
    expect(hasInnerWhitespace("  New Crawler ")).toBe(true);
    expect(visibleWhitespace("  New Crawler ")).toBe("new·crawler");
    expect(hasInnerWhitespace(" Ikman ")).toBe(false);
    expect(visibleWhitespace(" Ikman ")).toBe("ikman");
  });
});

describe("findExactNameMatch", () => {
  it("matches the saved name, ignoring case", () => {
    expect(findExactNameMatch(" Ikman ", sources)).toEqual({ name: "ikman" });
    expect(findExactNameMatch("DAILY NEWS", sources)).toEqual({ name: "daily news" });
  });

  it("still finds names saved before lowercasing", () => {
    expect(findExactNameMatch("topjobs", sources)).toEqual({ name: "TopJobs" });
  });

  it("finds nothing for a different name", () => {
    expect(findExactNameMatch("rooster", sources)).toBeUndefined();
  });
});

describe("confirmsSourceName", () => {
  it("needs the same name, ignoring case and outer spaces", () => {
    expect(confirmsSourceName("ikman", "ikman")).toBe(true);
    expect(confirmsSourceName(" Ikman ", "ikman")).toBe(true);
    expect(confirmsSourceName("topjobs", "TopJobs")).toBe(true);
    expect(confirmsSourceName("ikma", "ikman")).toBe(false);
  });
});
