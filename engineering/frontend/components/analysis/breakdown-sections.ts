import type { AnalysisSection } from "@/types/analysis";

interface BreakdownSectionLabel {
  title: string;
  /** Shown in the info tooltip next to the title. */
  tooltip?: string;
}

/**
 * Titles and tooltips for the analysis breakdown sections, shared by the chart view and the table
 * view so both show the same labels. Keyed like the API's sections. `satisfies` (not a type
 * annotation) keeps each entry's exact shape, so `.tooltip` is known to exist where it's set.
 */
export const BREAKDOWN_SECTIONS = {
  employment_sector: {
    title: "Employment Sector",
    tooltip: "Distribution of vacancies by employment sector",
  },
  experience: {
    title: "Experience",
    tooltip: "Distribution of vacancies by required years of experience",
  },
  province: {
    title: "Province",
    tooltip: "Distribution of vacancies by province",
  },
  education: {
    title: "Education Level",
    tooltip: "Distribution of vacancies by required education level",
  },
  formality: {
    title: "Formal / Informal",
    tooltip: "Proportion of formal vs informal sector vacancies",
  },
  gender: {
    title: "Gender",
    tooltip: "Gender preference specified in vacancy listings",
  },
  vocational_education: {
    title: "Vocational Education (NVQ Level)",
    tooltip: "Distribution of vacancies by NVQ vocational education level",
  },
  work_modes: { title: "Remote / On-Site / Hybrid" },
  job_type: { title: "Job Type" },
} satisfies Record<Exclude<AnalysisSection, "children" | "total_job_count">, BreakdownSectionLabel>;

/** Tooltip for the sub-classification breakdown; its title depends on the selected level. */
export const CHILDREN_BREAKDOWN_TOOLTIP =
  "Vacancy breakdown by sub-classification within the selected category";
