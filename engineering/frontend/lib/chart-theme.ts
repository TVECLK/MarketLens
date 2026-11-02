export const CHART_COLORS = {
  indigo: "#6366f1",
  teal: "#0d9488",
  sky: "#0284c7",
  violet: "#8b5cf6",
  amber: "#d97706",
  rose: "#f472b6",
  slate: "#94a3b8",
} as const;

export const GRID_STROKE = "#e4e4e7";
export const TICK_COLOR = "#71717a";

export const CHART_TOOLTIP_STYLE = {
  backgroundColor: "#18181b",
  border: "none",
  borderRadius: "8px",
  fontSize: "12px",
  color: "#fafafa",
};
export const CHART_TOOLTIP_ITEM_STYLE = { color: "#fafafa" };
export const CHART_TOOLTIP_LABEL_STYLE = { color: "#a1a1aa", fontWeight: 600 };
export const CHART_BAR_CURSOR = { fill: "#f4f4f5" };

// Palettes shared by every classification-hierarchy analysis view (industry,
// occupation, ...). Identical regardless of which hierarchy's primary/
// secondary accent is in play.
export const SECTOR_COLORS = [
  CHART_COLORS.sky,
  CHART_COLORS.teal,
  CHART_COLORS.indigo,
  CHART_COLORS.amber,
];
export const GENDER_COLORS = [CHART_COLORS.sky, CHART_COLORS.rose, CHART_COLORS.slate];

export interface HierarchyAccent {
  primary: string;
  secondary: string;
}

// Each classification hierarchy (industry, occupation) leads with its own
// primary/secondary accent; every other categorical color is shared.
export function buildCategoricalPalette(accent: HierarchyAccent): string[] {
  return [
    accent.primary,
    accent.secondary,
    CHART_COLORS.sky,
    CHART_COLORS.violet,
    CHART_COLORS.amber,
    CHART_COLORS.rose,
    CHART_COLORS.slate,
  ];
}

export function buildFormalInformalPalette(accent: HierarchyAccent): string[] {
  return [accent.primary, CHART_COLORS.slate];
}

export function buildProvincePalette(accent: HierarchyAccent): string[] {
  return [...buildCategoricalPalette(accent), "#0e7490", "#b45309"];
}
