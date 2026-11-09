// CATEGORY CONFIG — flat KPIs (Remote/On-Site and Contract Type intentionally
// removed per requirements — not present anywhere below).
export interface KPICategoryConfig {
  slug: string;
  title: string;
  description: string;
  itemLabel: string;
}

// The keys of this object are the category keys, and their order is the
// order categories appear in the nav.
export const KPI_CATEGORIES = {
  employmentSector: {
    slug: "employment-sector",
    title: "Employment Sector",
    description: "Manage emplyment sectors for job classification",
    itemLabel: "Employment Sector",
  },
  experience: {
    slug: "experience",
    title: "Experience",
    description: "Manage experiences for job classification",
    itemLabel: "Experience Level",
  },
  educationLevel: {
    slug: "education-level",
    title: "Education Level",
    description: "Manage education levels for job classification",
    itemLabel: "Education Level",
  },
  formalInformal: {
    slug: "formal-informal",
    title: "Formal / Informal",
    description: "Manage formality types for job classification",
    itemLabel: "Formality Type",
  },
  gender: {
    slug: "gender",
    title: "Gender",
    description: "Manage genders for job classification",
    itemLabel: "Gender",
  },
  vocationalEducation: {
    slug: "vocational-education",
    title: "Vocational Education (NVQ)",
    description: "Manage vocational education levels (NVQ) for job classification",
    itemLabel: "NVQ Level",
  },
} satisfies Record<string, KPICategoryConfig>;

export type KPICategoryKey = keyof typeof KPI_CATEGORIES;

export const CATEGORY_ORDER = Object.keys(KPI_CATEGORIES) as KPICategoryKey[];

// HIERARCHY CONFIG — Occupation + Industry, top level -> leaf level

export interface HierarchyLevelDef {
  key: string;
  /** Level name used by the analysis pages (URL param) and the backend analysis endpoints. */
  analysisLevel: string;
  /** Name of the level's list in the react-query cache. */
  queryKey: string;
  title: string;
  itemLabel: string;
}

export interface HierarchyDef {
  label: string;
  levels: readonly HierarchyLevelDef[];
}

// As with KPI_CATEGORIES, the keys are the hierarchy ids and their order is
// the nav order. This is the single place that names the levels of each
// hierarchy; the analysis pages and the KPI hooks derive theirs from here.
export const HIERARCHIES = {
  occupation: {
    label: "Occupation",
    levels: [
      { key: "majorGroup", analysisLevel: "major-group", queryKey: "major-groups", title: "Major Group", itemLabel: "Major Group" },
      { key: "subMajorGroup", analysisLevel: "sub-major-group", queryKey: "sub-major-groups", title: "Sub Major Group", itemLabel: "Sub Major Group" },
      { key: "minorGroup", analysisLevel: "minor-group", queryKey: "minor-groups", title: "Minor Group", itemLabel: "Minor Group" },
      { key: "unitGroup", analysisLevel: "unit-group", queryKey: "unit-groups", title: "Unit Group", itemLabel: "Unit Group" },
      { key: "occupationGroup", analysisLevel: "occupation-group", queryKey: "occupation-groups", title: "Occupation Group", itemLabel: "Occupation Group" },
    ],
  },
  industry: {
    label: "Industry",
    levels: [
      { key: "sector", analysisLevel: "industry-sector", queryKey: "industry-sectors", title: "Industry Sector", itemLabel: "Industry Sector" },
      { key: "division", analysisLevel: "industry-division", queryKey: "industry-divisions", title: "Industry Division", itemLabel: "Industry Division" },
      { key: "group", analysisLevel: "industry-group", queryKey: "industry-groups", title: "Industry Group", itemLabel: "Industry Group" },
      { key: "class", analysisLevel: "industry-class", queryKey: "industry-classes", title: "Industry Class", itemLabel: "Industry Class" },
      { key: "subClass", analysisLevel: "industry-subclass", queryKey: "industry-subclasses", title: "Industry Sub Class", itemLabel: "Industry Sub Class" },
    ],
  },
} as const satisfies Record<string, HierarchyDef>;

export type HierarchyId = keyof typeof HIERARCHIES;

export const HIERARCHY_ORDER = Object.keys(HIERARCHIES) as HierarchyId[];

export type OccupationLevel = (typeof HIERARCHIES.occupation.levels)[number]["analysisLevel"];
export type IndustryLevel = (typeof HIERARCHIES.industry.levels)[number]["analysisLevel"];

function levelLabels<L extends string>(
  levels: readonly { analysisLevel: L; title: string }[],
): Record<L, string> {
  return Object.fromEntries(levels.map((level) => [level.analysisLevel, level.title])) as Record<L, string>;
}

// Derived from HIERARCHIES, in top-to-leaf order.
export const OCCUPATION_LEVELS = HIERARCHIES.occupation.levels.map((level) => level.analysisLevel);
export const OCCUPATION_LEVEL_LABELS = levelLabels(HIERARCHIES.occupation.levels);
export const OCCUPATION_QUERY_KEYS = HIERARCHIES.occupation.levels.map((level) => level.queryKey);

export const INDUSTRY_LEVELS = HIERARCHIES.industry.levels.map((level) => level.analysisLevel);
export const INDUSTRY_LEVEL_LABELS = levelLabels(HIERARCHIES.industry.levels);
export const INDUSTRY_QUERY_KEYS = HIERARCHIES.industry.levels.map((level) => level.queryKey);

// ROUTING — each flat KPI lives at /kpis/<slug>, each hierarchy at
// /kpis/<hierarchyId>.
export function kpiCategoryHref(key: KPICategoryKey): string {
  return `/kpis/${KPI_CATEGORIES[key].slug}`;
}

export function hierarchyHref(id: HierarchyId): string {
  return `/kpis/${id}`;
}

export function categoryKeyFromSlug(slug: string): KPICategoryKey | null {
  return CATEGORY_ORDER.find((key) => KPI_CATEGORIES[key].slug === slug) ?? null;
}
