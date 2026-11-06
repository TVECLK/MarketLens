import type { SectionError } from "@/types/bff";

export interface VacancyTrendPoint {
  label: string;
  open_job_count: number;
}

export interface VacancyTrendResponse {
  granularity: "daily" | "monthly";
  count: number;
  data: VacancyTrendPoint[];
}

export interface VacancyTotalResponse {
  total_vacancies: number;
}

export interface RootJobCountStat {
  id: number;
  name: string;
  open_job_count: number;
}

export interface VacancySummaryResponse {
  count: number;
  data: RootJobCountStat[];
}

export type DashboardSection =
  | "vacancy_trend"
  | "vacancy_total"
  | "by_occupation"
  | "by_industry";

// A section is `null` when its backend call failed; the reason is in `errors`.
export interface DashboardOverview {
  from_date: string | null;
  to_date: string | null;
  vacancy_trend: VacancyTrendResponse | null;
  vacancy_total: VacancyTotalResponse | null;
  by_occupation: VacancySummaryResponse | null;
  by_industry: VacancySummaryResponse | null;
  errors: Partial<Record<DashboardSection, SectionError>>;
}