import { describe, it, expect, vi, afterEach } from "vitest";
import { toISODate, getDataRangeStart, getDataRangeEnd } from "@/lib/date-range";

afterEach(() => {
  vi.useRealTimers();
});

describe("toISODate", () => {
  it("formats a date as YYYY-MM-DD", () => {
    expect(toISODate(new Date(Date.UTC(2026, 2, 7)))).toBe("2026-03-07");
  });

  it("zero-pads single-digit months and days", () => {
    expect(toISODate(new Date(Date.UTC(2026, 0, 5)))).toBe("2026-01-05");
  });
});

describe("getDataRangeEnd", () => {
  it("reflects the current system time, not a frozen value", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2020, 5, 15));
    const first = getDataRangeEnd();

    vi.setSystemTime(new Date(2024, 8, 1));
    const second = getDataRangeEnd();

    // This is the regression this module exists to prevent: if "today" were
    // ever cached in a module-level constant instead of computed per call,
    // these two would be equal even though the clock moved.
    expect(first.getTime()).not.toBe(second.getTime());
    expect(second.getFullYear()).toBe(2024);
  });
});

describe("getDataRangeStart", () => {
  it("is January 1st, 3 years before getDataRangeEnd's year", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 5, 15));

    const start = getDataRangeStart();

    expect(start.getFullYear()).toBe(2023);
    expect(start.getMonth()).toBe(0);
    expect(start.getDate()).toBe(1);
  });
});
