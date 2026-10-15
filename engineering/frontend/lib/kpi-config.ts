// CATEGORY CONFIG — flat KPIs (Remote/On-Site and Contract Type intentionally
// removed per requirements — not present anywhere below).
export type KPICategoryKey =
  | "employmentSector"
  | "experience"
  | "educationLevel"
  | "formalInformal"
  | "gender"
  | "vocationalEducation";

export interface KPICategoryConfig {
  title: string;
  description: string;
  itemLabel: string;
  accent: string;
}

export const KPI_CATEGORIES: Record<KPICategoryKey, KPICategoryConfig> = {
  employmentSector: {
    title: "Employment Sector",
    description: "Government, Semi-Government, Private and NGO sector breakdown",
    itemLabel: "Sector",
    accent: "blue",
  },
  experience: {
    title: "Experience",
    description: "Experience breakdown used for job categrization",
    itemLabel: "Experience Level",
    accent: "amber",
  },
  educationLevel: {
    title: "Education Level",
    description: "Academic qualification levels",
    itemLabel: "Education Level",
    accent: "violet",
  },
  formalInformal: {
    title: "Formal / Informal",
    description: "Formal vs. informal sector classification",
    itemLabel: "Type",
    accent: "emerald",
  },
  gender: {
    title: "Gender",
    description: "Gender breakdown of the vacancies",
    itemLabel: "Gender",
    accent: "pink",
  },
  vocationalEducation: {
    title: "Vocational Education (NVQ)",
    description: "National Vocational Qualification levels",
    itemLabel: "NVQ Level",
    accent: "indigo",
  },
};

export const CATEGORY_ORDER: KPICategoryKey[] = [
  "employmentSector",
  "experience",
  "educationLevel",
  "formalInformal",
  "gender",
  "vocationalEducation",
];

// HIERARCHY CONFIG — Occupation + Industry, top level -> leaf level
export type HierarchyId = "occupation" | "industry";

export interface HierarchyLevelDef {
  key: string;
  title: string;
  itemLabel: string;
}

export interface HierarchyDef {
  label: string;
  accent: string;
  levels: HierarchyLevelDef[];
}

export const HIERARCHY_ORDER: HierarchyId[] = ["occupation", "industry"];

export const HIERARCHIES: Record<HierarchyId, HierarchyDef> = {
  occupation: {
    label: "Occupation",
    accent: "rose",
    levels: [
      { key: "majorGroup", title: "Major Group", itemLabel: "Major Group" },
      { key: "subMajorGroup", title: "Sub Major Group", itemLabel: "Sub Major Group" },
      { key: "minorGroup", title: "Minor Group", itemLabel: "Minor Group" },
      { key: "unitGroup", title: "Unit Group", itemLabel: "Unit Group" },
      { key: "occupationGroup", title: "Occupation Group", itemLabel: "Occupation Group" },
    ],
  },
  industry: {
    label: "Industry",
    accent: "orange",
    levels: [
      { key: "sector", title: "Industry Sector", itemLabel: "Industry Sector" },
      { key: "division", title: "Industry Division", itemLabel: "Industry Division" },
      { key: "group", title: "Industry Group", itemLabel: "Industry Group" },
      { key: "class", title: "Industry Class", itemLabel: "Industry Class" },
      { key: "subClass", title: "Industry Sub Class", itemLabel: "Industry Sub Class" },
    ],
  },
};

// ACCENT COLOR CLASS MAP (Tailwind needs literal class names, not template strings)
export interface AccentClasses {
  bg: string;
  text: string;
  ring: string;
  solidBg: string;
  lightBg: string;
}

export const ACCENT_CLASSES: Record<string, AccentClasses> = {
  blue: { bg: "bg-blue-600", text: "text-blue-600", ring: "ring-blue-500", solidBg: "bg-blue-600", lightBg: "bg-blue-50" },
  amber: { bg: "bg-amber-600", text: "text-amber-600", ring: "ring-amber-500", solidBg: "bg-amber-500", lightBg: "bg-amber-50" },
  violet: { bg: "bg-violet-600", text: "text-violet-600", ring: "ring-violet-500", solidBg: "bg-violet-600", lightBg: "bg-violet-50" },
  emerald: { bg: "bg-emerald-600", text: "text-emerald-600", ring: "ring-emerald-500", solidBg: "bg-emerald-600", lightBg: "bg-emerald-50" },
  pink: { bg: "bg-pink-600", text: "text-pink-600", ring: "ring-pink-500", solidBg: "bg-pink-600", lightBg: "bg-pink-50" },
  indigo: { bg: "bg-indigo-600", text: "text-indigo-600", ring: "ring-indigo-500", solidBg: "bg-indigo-600", lightBg: "bg-indigo-50" },
  rose: { bg: "bg-rose-600", text: "text-rose-600", ring: "ring-rose-500", solidBg: "bg-rose-600", lightBg: "bg-rose-50" },
  orange: { bg: "bg-orange-600", text: "text-orange-600", ring: "ring-orange-500", solidBg: "bg-orange-600", lightBg: "bg-orange-50" },
};
