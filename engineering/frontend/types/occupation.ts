import type { AnalysisResponse, HierarchyNode } from "./analysis";

export type { HierarchyNode } from "./analysis";

export interface MajorGroupListResponse {
  count: number;
  major_groups: HierarchyNode[];
}
export interface SubMajorGroupChildrenResponse {
  major_group_id: number;
  count: number;
  sub_major_groups: HierarchyNode[];
}
export interface MinorGroupChildrenResponse {
  sub_major_group_id: number;
  count: number;
  minor_groups: HierarchyNode[];
}
export interface UnitGroupChildrenResponse {
  minor_group_id: number;
  count: number;
  unit_groups: HierarchyNode[];
}
export interface OccupationGroupChildrenResponse {
  unit_group_id: number;
  count: number;
  occupation_groups: HierarchyNode[];
}

export type OccupationAnalysisResponse = AnalysisResponse;

export interface SkillStat { id: number; skill: string; open_job_count: number; }
export interface EmployerStat { id: number; name: string; open_job_count: number; }

export interface Top15SkillsResponse {
  level: string; id: number; from_date: string; to_date: string;
  count: number; skills: SkillStat[];
}
export interface AllSkillsResponse {
  level: string; id: number; from_date: string; to_date: string;
  count: number; total: number; limit: number; offset: number; skills: SkillStat[];
}
export interface TopHiringEmployersResponse {
  level: string; id: number; from_date: string; to_date: string;
  count: number; employers: EmployerStat[];
}
export interface OccupationSkillsResponse {
  level: string; id: string; from_date: string | null; to_date: string | null;
  top_15_skills: Top15SkillsResponse;
  all_skills: AllSkillsResponse;
  top_hiring_employers: TopHiringEmployersResponse;
}
