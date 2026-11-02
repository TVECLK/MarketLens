import { describe, it, expect, vi, afterEach } from "vitest";
import { toISODate, getMinSelectableDate, getMaxSelectableDate } from "@/lib/date-range";

afterEach(() => {
  vi.useRealTimers();
});

describe("toISODate", () => {
  it("formats a date as YYYY-MM-DD", () => {
    expect(toISODate(new Date(2026, 2, 7))).toBe("2026-03-07");
  });

  it("zero-pads single-digit months and days", () => {
    expect(toISODate(new Date(2026, 0, 5))).toBe("2026-01-05");
  });

  it("uses the local calendar day, not UTC", () => {
    expect(toISODate(new Date(2026, 0, 1))).toBe("2026-01-01");
    expect(toISODate(new Date(2026, 5, 15, 23, 30))).toBe("2026-06-15");
  });
});

describe("getMaxSelectableDate", () => {
  it("reflects the current system time, not a frozen value", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2020, 5, 15));
    const first = getMaxSelectableDate();

    vi.setSystemTime(new Date(2024, 8, 1));
    const second = getMaxSelectableDate();

    // This is the regression this module exists to prevent: if "today" were
    // ever cached in a module-level constant instead of computed per call,
    // these two would be equal even though the clock moved.
    expect(first.getTime()).not.toBe(second.getTime());
    expect(second.getFullYear()).toBe(2024);
  });
});

describe("getMinSelectableDate", () => {
  it("is January 1st, 3 years before getMaxSelectableDate's year", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 5, 15));

    const start = getMinSelectableDate();

    expect(start.getFullYear()).toBe(2023);
    expect(start.getMonth()).toBe(0);
    expect(start.getDate()).toBe(1);
  });
});
