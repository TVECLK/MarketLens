import { describe, it, expect } from "vitest";
import {
  CHART_COLORS,
  buildCategoricalPalette,
  buildFormalInformalPalette,
  buildProvincePalette,
  type HierarchyAccent,
} from "@/lib/chart-theme";

const industryAccent: HierarchyAccent = {
  primary: CHART_COLORS.teal,
  secondary: CHART_COLORS.indigo,
};

describe("buildCategoricalPalette", () => {
  it("leads with the accent's primary and secondary colors", () => {
    const palette = buildCategoricalPalette(industryAccent);
    expect(palette[0]).toBe(CHART_COLORS.teal);
    expect(palette[1]).toBe(CHART_COLORS.indigo);
  });

  it("always returns 7 colors", () => {
    expect(buildCategoricalPalette(industryAccent)).toHaveLength(7);
  });
});

describe("buildFormalInformalPalette", () => {
  it("returns exactly [primary, slate]", () => {
    expect(buildFormalInformalPalette(industryAccent)).toEqual([
      CHART_COLORS.teal,
      CHART_COLORS.slate,
    ]);
  });
});

describe("buildProvincePalette", () => {
  it("extends the categorical palette with 2 extra colors", () => {
    const categorical = buildCategoricalPalette(industryAccent);
    const province = buildProvincePalette(industryAccent);
    expect(province.slice(0, categorical.length)).toEqual(categorical);
    expect(province).toHaveLength(categorical.length + 2);
  });
});
