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

export interface DashboardOverview {
  vacancy_trend: VacancyTrendResponse;
  vacancy_total: VacancyTotalResponse;
  by_occupation: OccupationsByDateRangeResponse;
  by_industry: IndustriesByDateRangeResponse;
}