"use client";

import { useMemo } from "react";
import { useDashboardOverview } from "@/hooks/use-dashboard";
import { usePersistedDateRange } from "@/hooks/use-persisted-date-range";
import { useIsMobile } from "@/hooks/use-is-mobile";
import { useScrolled } from "@/hooks/use-scrolled";
import { getDataRangeStart, getDataRangeEnd } from "@/lib/date-range";
import Header from "@/components/layout/Header";
import DateRangePicker from "@/components/shared/DateRangePicker";
import VacancyTrendChart from "@/components/dashboard/VacancyTrendChart";
import OccupationDistributionChart from "@/components/dashboard/OccupationDistributionChart";
import IndustryDistributionChart from "@/components/dashboard/IndustryDistributionChart";

const DATE_RANGE_STORAGE_KEY = "lmis-date-range:v2";

export default function DashboardPage() {
  const { fromDate, toDate, setFromDate, setToDate, rangeInvalid } =
    usePersistedDateRange(DATE_RANGE_STORAGE_KEY);
  const isMobile = useIsMobile();
  const scrolled = useScrolled();
  const dataRangeEnd = getDataRangeEnd();
  const dataRangeStart = getDataRangeStart();

  const {
    data: overview,
    isLoading,
    isError,
    error,
  } = useDashboardOverview(fromDate, toDate);

  const trendData = useMemo(() => overview?.vacancy_trend.data ?? [], [overview]);
  const granularity = overview?.vacancy_trend.granularity ?? "weekly";
  const totalInRange = overview?.vacancy_total.total_vacancies ?? 0;

  const occData = useMemo(
    () =>
      (overview?.by_occupation.occupations ?? []).map((o) => ({
        name: o.name,
        value: o.open_job_count,
      })),
    [overview],
  );

  const indData = useMemo(
    () =>
      (overview?.by_industry.industries ?? []).map((i) => ({
        name: i.name,
        value: i.open_job_count,
      })),
    [overview],
  );

  return (
    <div className="min-h-screen w-full min-w-0 bg-zinc-50 text-zinc-800 font-sans">
      <Header
        title="Labour Market Demand Dashboard"
        subtitle="National overview of labour market demand across occupations and industries"
        logoSrc="/logo.png"
      />

      <div className="p-4 md:p-8 space-y-6 md:space-y-8 w-full max-w-[1400px] mx-auto pb-20">
        <DateRangePicker
          fromDate={fromDate}
          toDate={toDate}
          onFromDateChange={setFromDate}
          onToDateChange={setToDate}
          minDate={dataRangeStart}
          maxDate={dataRangeEnd}
          scrolled={scrolled}
          description="Select a date range to view labour market analytics for that period"
        />

        {isMobile && (
          <div className="md:hidden bg-indigo-50 border border-indigo-200 text-indigo-700 text-xs font-medium px-4 py-2 rounded-xl flex items-center gap-2">
            💡 Rotate your device to landscape for a better chart view
          </div>
        )}

        {rangeInvalid && (
          <div className="bg-zinc-100 border border-zinc-300 text-zinc-700 text-xs font-bold px-4 py-3 rounded-xl">
            The start date must be before the end date.
          </div>
        )}

        {isError && !rangeInvalid && (
          <div className="bg-red-50 border border-red-200 text-red-700 text-xs font-bold px-4 py-3 rounded-xl">
            Failed to load dashboard data
            {error instanceof Error ? `: ${error.message}` : "."}
          </div>
        )}

        <VacancyTrendChart
          data={trendData}
          granularity={granularity}
          totalVacancies={totalInRange}
          isLoading={isLoading}
        />

        <OccupationDistributionChart
          data={occData}
          isMobile={isMobile}
          isLoading={isLoading}
          linkHref={`/occupations?from=${fromDate}&to=${toDate}`}
        />

        <IndustryDistributionChart
          data={indData}
          isLoading={isLoading}
          linkHref={`/industries?from=${fromDate}&to=${toDate}`}
        />
      </div>
    </div>
  );
}
