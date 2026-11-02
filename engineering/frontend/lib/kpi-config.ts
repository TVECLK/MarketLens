// CATEGORY CONFIG — flat KPIs (Remote/On-Site and Contract Type intentionally
// removed per requirements — not present anywhere below).
export interface KPICategoryConfig {
  slug: string;
  title: string;
  description: string;
  itemLabel: string;
  accent: string;
}

// The keys of this object are the category keys, and their order is the
// order categories appear in the nav.
export const KPI_CATEGORIES = {
  employmentSector: {
    slug: "employment-sector",
    title: "Employment Sector",
    description: "Government, Semi-Government, Private and NGO sector breakdown",
    itemLabel: "Sector",
    accent: "blue",
  },
  experience: {
    slug: "experience",
    title: "Experience",
    description: "Experience breakdown used for job categorization",
    itemLabel: "Experience Level",
    accent: "amber",
  },
  educationLevel: {
    slug: "education-level",
    title: "Education Level",
    description: "Academic qualification levels",
    itemLabel: "Education Level",
    accent: "violet",
  },
  formalInformal: {
    slug: "formal-informal",
    title: "Formal / Informal",
    description: "Formal vs. informal sector classification",
    itemLabel: "Type",
    accent: "emerald",
  },
  gender: {
    slug: "gender",
    title: "Gender",
    description: "Gender breakdown of the vacancies",
    itemLabel: "Gender",
    accent: "pink",
  },
  vocationalEducation: {
    slug: "vocational-education",
    title: "Vocational Education (NVQ)",
    description: "National Vocational Qualification levels",
    itemLabel: "NVQ Level",
    accent: "indigo",
  },
} satisfies Record<string, KPICategoryConfig>;

export type KPICategoryKey = keyof typeof KPI_CATEGORIES;

export const CATEGORY_ORDER = Object.keys(KPI_CATEGORIES) as KPICategoryKey[];

// HIERARCHY CONFIG — Occupation + Industry, top level -> leaf level

export interface HierarchyLevelDef {
  key: string;
  slug: string;
  title: string;
  itemLabel: string;
}

export interface HierarchyDef {
  label: string;
  accent: string;
  levels: HierarchyLevelDef[];
}

// As with KPI_CATEGORIES, the keys are the hierarchy ids and their order is
// the nav order.
export const HIERARCHIES = {
  occupation: {
    label: "Occupation",
    accent: "rose",
    levels: [
      { key: "majorGroup", slug: "major-group", title: "Major Group", itemLabel: "Major Group" },
      { key: "subMajorGroup", slug: "sub-major-group", title: "Sub Major Group", itemLabel: "Sub Major Group" },
      { key: "minorGroup", slug: "minor-group", title: "Minor Group", itemLabel: "Minor Group" },
      { key: "unitGroup", slug: "unit-group", title: "Unit Group", itemLabel: "Unit Group" },
      { key: "occupationGroup", slug: "occupation-group", title: "Occupation Group", itemLabel: "Occupation Group" },
    ],
  },
  industry: {
    label: "Industry",
    accent: "orange",
    levels: [
      { key: "sector", slug: "sector", title: "Industry Sector", itemLabel: "Industry Sector" },
      { key: "division", slug: "division", title: "Industry Division", itemLabel: "Industry Division" },
      { key: "group", slug: "group", title: "Industry Group", itemLabel: "Industry Group" },
      { key: "class", slug: "class", title: "Industry Class", itemLabel: "Industry Class" },
      { key: "subClass", slug: "sub-class", title: "Industry Sub Class", itemLabel: "Industry Sub Class" },
    ],
  },
} satisfies Record<string, HierarchyDef>;

export type HierarchyId = keyof typeof HIERARCHIES;

export const HIERARCHY_ORDER = Object.keys(HIERARCHIES) as HierarchyId[];

// ROUTING — each flat KPI lives at /kpis/<slug>, each hierarchy level at
// /kpis/<hierarchyId>/<level slug>.
export function kpiCategoryHref(key: KPICategoryKey): string {
  return `/kpis/${KPI_CATEGORIES[key].slug}`;
}

export function hierarchyLevelHref(id: HierarchyId, levelIndex: number): string {
  return `/kpis/${id}/${HIERARCHIES[id].levels[levelIndex].slug}`;
}

export function categoryKeyFromSlug(slug: string): KPICategoryKey | null {
  return CATEGORY_ORDER.find((key) => KPI_CATEGORIES[key].slug === slug) ?? null;
}

/** Returns -1 when the slug isn't a level of this hierarchy. */
export function levelIndexFromSlug(id: HierarchyId, slug: string): number {
  return HIERARCHIES[id].levels.findIndex((level) => level.slug === slug);
}

// ACCENT COLOR CLASS MAP (Tailwind needs literal class names, not template strings)
export interface AccentClasses {
  bg: string;
  text: string;
  ring: string;
  solidBg: string;
  /** Text color for content placed on `bg` / `solidBg` (pastel fills need dark text). */
  solidText: string;
  lightBg: string;
}

export const ACCENT_CLASSES: Record<string, AccentClasses> = {
  blue: { bg: "bg-blue-200", text: "text-blue-700", ring: "ring-blue-200", solidBg: "bg-blue-100", solidText: "text-blue-800", lightBg: "bg-blue-50" },
  amber: { bg: "bg-amber-200", text: "text-amber-700", ring: "ring-amber-200", solidBg: "bg-amber-100", solidText: "text-amber-800", lightBg: "bg-amber-50" },
  violet: { bg: "bg-violet-200", text: "text-violet-700", ring: "ring-violet-200", solidBg: "bg-violet-100", solidText: "text-violet-800", lightBg: "bg-violet-50" },
  emerald: { bg: "bg-emerald-200", text: "text-emerald-700", ring: "ring-emerald-200", solidBg: "bg-emerald-100", solidText: "text-emerald-800", lightBg: "bg-emerald-50" },
  pink: { bg: "bg-pink-200", text: "text-pink-700", ring: "ring-pink-200", solidBg: "bg-pink-100", solidText: "text-pink-800", lightBg: "bg-pink-50" },
  indigo: { bg: "bg-indigo-200", text: "text-indigo-700", ring: "ring-indigo-200", solidBg: "bg-indigo-100", solidText: "text-indigo-800", lightBg: "bg-indigo-50" },
  rose: { bg: "bg-rose-200", text: "text-rose-700", ring: "ring-rose-200", solidBg: "bg-rose-100", solidText: "text-rose-800", lightBg: "bg-rose-50" },
  orange: { bg: "bg-orange-200", text: "text-orange-700", ring: "ring-orange-200", solidBg: "bg-orange-100", solidText: "text-orange-800", lightBg: "bg-orange-50" },
};
