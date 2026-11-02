// Shared shapes for the industry and occupation classification hierarchies.
// Both standards (SLSIC and SLSO) expose structurally identical analysis
// payloads — only the concrete values differ — so one definition backs both
// `types/industry.ts` and `types/occupation.ts`.

import type { SectionError } from "@/types/bff";

export interface HierarchyNode {
  id: number;
  name: string;
  code: string;
}

export interface TotalJobCountResult {
  standard: string;
  level: string;
  id: number;
  total_job_count: number;
}

export interface ChildJobCount {
  id: number;
  name: string;
  code: string;
  open_job_count: number;
}

export interface ChildrenResult {
  standard: string;
  level: string;
  id: number;
  child_level: string;
  count: number;
  children: ChildJobCount[];
}

export interface EmploymentSectorStat { id: number; sector: string; open_job_count: number; }
export interface ExperienceStat { id: number; name: string; open_job_count: number; }
export interface ProvinceStat { id: number; province: string; lat: number; lng: number; open_job_count: number; }
export interface EducationLevelStat { id: number; level: string; open_job_count: number; }
export interface FormalityStat { id: number; formality_type: string; open_job_count: number; }
export interface GenderStat { id: number; gender_type: string; open_job_count: number; }
export interface VocationalEducationStat { id: number; level: string; open_job_count: number; }
export interface JobTypeStat { id: number; type: string; open_job_count: number; }
export interface RemoteOnSiteHybridStat { remote_count: number; on_site_count: number; hybrid_count: number; }

// Sections that can fail on their own. `children` is absent: it legitimately
// fails on leaf levels, so the backend never reports it as an error.
export type AnalysisSection =
  | "total_job_count"
  | "employment_sector"
  | "experience"
  | "province"
  | "education"
  | "formality"
  | "gender"
  | "vocational_education"
  | "remote_onsite_hybrid"
  | "job_type";

export type AnalysisSectionErrors = Partial<Record<AnalysisSection, SectionError>>;

// A section is `null` when its backend call failed; the reason is in `errors`.
export interface AnalysisResponse {
  standard: string;
  level: string;
  id: string;
  from_date: string | null;
  to_date: string | null;
  total_job_count: TotalJobCountResult | null;
  children: ChildrenResult | null;
  employment_sector: { employment_sectors: EmploymentSectorStat[] } | null;
  experience: { experiences: ExperienceStat[] } | null;
  province: { provinces: ProvinceStat[] } | null;
  education: { education_levels: EducationLevelStat[] } | null;
  formality: { formalities: FormalityStat[] } | null;
  gender: { genders: GenderStat[] } | null;
  vocational_education: { vocational_educations: VocationalEducationStat[] } | null;
  remote_onsite_hybrid: { remote_vs_onsite_vs_hybrid: RemoteOnSiteHybridStat } | null;
  job_type: { job_types: JobTypeStat[] } | null;
  errors: AnalysisSectionErrors;
}
