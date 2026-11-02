import { describe, expect, it } from "vitest";
import { isEmptyChartData } from "./chart-data";

const value = (d: { value: number }) => d.value;

describe("isEmptyChartData", () => {
  it("is true for an empty list", () => {
    expect(isEmptyChartData([], value)).toBe(true);
  });

  it("is true when every value is zero", () => {
    expect(isEmptyChartData([{ value: 0 }, { value: 0 }], value)).toBe(true);
  });

  it("is false when any value is non-zero", () => {
    expect(isEmptyChartData([{ value: 0 }, { value: 3 }], value)).toBe(false);
  });

  it("reads the value through the accessor", () => {
    expect(isEmptyChartData([{ open_job_count: 0 }], (p) => p.open_job_count)).toBe(true);
    expect(isEmptyChartData([{ open_job_count: 2 }], (p) => p.open_job_count)).toBe(false);
  });
});
