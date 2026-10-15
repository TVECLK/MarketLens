// Shared shapes for the industry and occupation classification hierarchies.
// Both standards (SLSIC and SLSO) expose structurally identical analysis
// payloads — only the concrete values differ — so one definition backs both
// `types/industry.ts` and `types/occupation.ts`.

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

export interface AnalysisResponse {
  standard: string;
  level: string;
  id: string;
  from_date: string | null;
  to_date: string | null;
  total_job_count: TotalJobCountResult;
  children: ChildrenResult | null;
  employment_sector: { employment_sectors: EmploymentSectorStat[] };
  experience: { experiences: ExperienceStat[] };
  province: { provinces: ProvinceStat[] };
  education: { education_levels: EducationLevelStat[] };
  formality: { formalities: FormalityStat[] };
  gender: { genders: GenderStat[] };
  vocational_education: { vocational_educations: VocationalEducationStat[] };
  remote_onsite_hybrid: { remote_vs_onsite_vs_hybrid: RemoteOnSiteHybridStat };
  job_type: { job_types: JobTypeStat[] };
}
