"use client";

import { useMemo } from "react";
import { useDashboardOverview } from "@/hooks/use-dashboard";
import { useDateRange } from "@/providers/date-range-provider";
import type { DashboardSection } from "@/types/dashboard";
import VacancyTrendChart from "@/components/dashboard/VacancyTrendChart";
import OccupationDistributionChart from "@/components/dashboard/OccupationDistributionChart";
import IndustryDistributionChart from "@/components/dashboard/IndustryDistributionChart";

export default function DashboardPage() {
  const { fromDate, toDate } = useDateRange();

  const {
    data: overview,
    isLoading,
    isError,
    error,
  } = useDashboardOverview(fromDate, toDate);

  // A chart shows its error when the whole request failed or only its own
  // section did, so the sections that loaded are still drawn.
  const sectionFailed = (section: DashboardSection) =>
    isError || Boolean(overview?.errors[section]);

  const trendData = useMemo(() => overview?.vacancy_trend?.data ?? [], [overview]);
  const granularity = overview?.vacancy_trend?.granularity;
  const totalInRange = overview?.vacancy_total?.total_vacancies;

  const occData = useMemo(
    () =>
      (overview?.by_occupation?.occupations ?? []).map((o) => ({
        name: o.name,
        value: o.open_job_count,
      })),
    [overview],
  );

  const indData = useMemo(
    () =>
      (overview?.by_industry?.industries ?? []).map((i) => ({
        name: i.name,
        value: i.open_job_count,
      })),
    [overview],
  );

  return (
    <div className="p-4 md:p-8 space-y-6 md:space-y-8 w-full pb-20">
      {isError && (
        <div className="bg-red-50 border border-red-200 text-red-700 text-xs font-bold px-4 py-3 rounded-md">
          Failed to load dashboard data
          {error instanceof Error ? `: ${error.message}` : "."}
        </div>
      )}

      <VacancyTrendChart
        data={trendData}
        granularity={granularity}
        totalVacancies={totalInRange}
        isLoading={isLoading}
        isError={sectionFailed("vacancy_trend")}
      />

      <OccupationDistributionChart
        data={occData}
        isLoading={isLoading}
        isError={sectionFailed("by_occupation")}
        linkHref={`/occupations?from=${fromDate}&to=${toDate}`}
      />

      <IndustryDistributionChart
        data={indData}
        isLoading={isLoading}
        isError={sectionFailed("by_industry")}
        linkHref={`/industries?from=${fromDate}&to=${toDate}`}
      />
    </div>
  );
}
