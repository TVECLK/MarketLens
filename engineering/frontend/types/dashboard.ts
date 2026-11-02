import type { SectionError } from "@/types/bff";

export interface VacancyTrendPoint {
  label: string;
  open_job_count: number;
}

export interface VacancyTrendResponse {
  from_date: string;
  to_date: string;
  granularity: "weekly" | "monthly";
  count: number;
  data: VacancyTrendPoint[];
}

export interface VacancyTotalResponse {
  from_date: string;
  to_date: string;
  total_vacancies: number;
}

export interface OccupationDateRangeStat {
  id: number;
  name: string;
  open_job_count: number;
}

export interface OccupationsByDateRangeResponse {
  from_date: string;
  to_date: string;
  count: number;
  occupations: OccupationDateRangeStat[];
}

export interface IndustryDateRangeStat {
  id: number;
  name: string;
  open_job_count: number;
}

export interface IndustriesByDateRangeResponse {
  from_date: string;
  to_date: string;
  count: number;
  industries: IndustryDateRangeStat[];
}

export type DashboardSection =
  | "vacancy_trend"
  | "vacancy_total"
  | "by_occupation"
  | "by_industry";

// A section is `null` when its backend call failed; the reason is in `errors`.
export interface DashboardOverview {
  vacancy_trend: VacancyTrendResponse | null;
  vacancy_total: VacancyTotalResponse | null;
  by_occupation: OccupationsByDateRangeResponse | null;
  by_industry: IndustriesByDateRangeResponse | null;
  errors: Partial<Record<DashboardSection, SectionError>>;
}