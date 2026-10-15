"use client";

import { Suspense, useMemo, useState } from "react";
import Header from "@/components/layout/Header";
import BackLink from "@/components/shared/BackLink";
import DateRangePicker from "@/components/shared/DateRangePicker";
import ClassificationHierarchyPicker from "@/components/analysis/ClassificationHierarchyPicker";
import AnalysisHeaderBar from "@/components/analysis/AnalysisHeaderBar";
import AnalysisBreakdownSection, {
  type AnalysisViewMode,
} from "@/components/analysis/AnalysisBreakdownSection";
import { useScrolled } from "@/hooks/use-scrolled";
import { useUrlAwareDateRange } from "@/hooks/use-url-aware-date-range";
import {
  useHierarchyPath,
  useHierarchyDrilldown,
  type HierarchyLevelState,
} from "@/hooks/use-hierarchy-drilldown";
import { useAnalysisBreakdown } from "@/hooks/use-analysis-breakdown";
import {
  useIndustrySectors,
  useIndustryDivisions,
  useIndustryGroups,
  useIndustryClasses,
  useIndustrySubclasses,
  useIndustryAnalysis,
} from "@/hooks/use-industry";
import type { HierarchyNode } from "@/types/analysis";
import { CHART_COLORS } from "@/lib/chart-theme";
import { getDataRangeStart, getDataRangeEnd } from "@/lib/date-range";

const DATE_RANGE_STORAGE_KEY = "lmis-date-range:v2";

type IndustryLevel =
  | "industry-sector"
  | "industry-division"
  | "industry-group"
  | "industry-class"
  | "industry-subclass";

const LEVELS: IndustryLevel[] = [
  "industry-sector",
  "industry-division",
  "industry-group",
  "industry-class",
  "industry-subclass",
];
const LEVEL_LABELS: Record<IndustryLevel, string> = {
  "industry-sector": "Industry Sector",
  "industry-division": "Industry Division",
  "industry-group": "Industry Group",
  "industry-class": "Industry Class",
  "industry-subclass": "Industry Subclass",
};
const MAX_LEVELS = LEVELS.length;

const INDUSTRY_ACCENT = { primary: CHART_COLORS.teal, secondary: CHART_COLORS.indigo };

function IndustryAnalysis() {
  const { fromDate, toDate, setFromDate, setToDate, rangeInvalid } =
    useUrlAwareDateRange(DATE_RANGE_STORAGE_KEY);
  const scrolled = useScrolled();
  const [viewMode, setViewMode] = useState<AnalysisViewMode>("chart");
  const dataRangeEnd = getDataRangeEnd();
  const dataRangeStart = getDataRangeStart();

  const [path, setPath] = useHierarchyPath<IndustryLevel>();

  const industrySectorsQuery = useIndustrySectors(fromDate, toDate);
  const industrySectors = useMemo<HierarchyNode[]>(
    () =>
      (industrySectorsQuery.data?.industry_sectors ?? []).map((s) => ({
        id: s.id,
        name: s.name,
        code: "",
      })),
    [industrySectorsQuery.data],
  );

  const industryDivisionsQuery = useIndustryDivisions(path[0]?.id ?? null, fromDate, toDate);
  const industryGroupsQuery = useIndustryGroups(path[1]?.id ?? null, fromDate, toDate);
  const industryClassesQuery = useIndustryClasses(path[2]?.id ?? null, fromDate, toDate);
  const industrySubclassesQuery = useIndustrySubclasses(path[3]?.id ?? null, fromDate, toDate);

  const levels: HierarchyLevelState<IndustryLevel>[] = [
    {
      key: "industry-sector",
      label: LEVEL_LABELS["industry-sector"],
      options: industrySectors,
      isLoading: industrySectorsQuery.isLoading,
      isSuccess: industrySectorsQuery.isSuccess,
    },
    {
      key: "industry-division",
      label: LEVEL_LABELS["industry-division"],
      options: industryDivisionsQuery.data?.industry_divisions ?? [],
      isLoading: industryDivisionsQuery.isLoading,
      isSuccess: industryDivisionsQuery.isSuccess,
    },
    {
      key: "industry-group",
      label: LEVEL_LABELS["industry-group"],
      options: industryGroupsQuery.data?.industry_groups ?? [],
      isLoading: industryGroupsQuery.isLoading,
      isSuccess: industryGroupsQuery.isSuccess,
    },
    {
      key: "industry-class",
      label: LEVEL_LABELS["industry-class"],
      options: industryClassesQuery.data?.industry_classes ?? [],
      isLoading: industryClassesQuery.isLoading,
      isSuccess: industryClassesQuery.isSuccess,
    },
    {
      key: "industry-subclass",
      label: LEVEL_LABELS["industry-subclass"],
      options: industrySubclassesQuery.data?.industry_subclasses ?? [],
      isLoading: industrySubclassesQuery.isLoading,
      isSuccess: industrySubclassesQuery.isSuccess,
    },
  ];

  const { prunedNotice, handleSelect, handleAddLevel, handleRemoveLevel } =
    useHierarchyDrilldown(path, setPath, levels);

  const deepest = path[path.length - 1] ?? null;
  const deepestLevelIndex = path.length - 1;
  const isInitializing = industrySectorsQuery.isLoading;

  const {
    data: analysis,
    isLoading: analysisLoading,
    isError: analysisError,
    error: analysisErrorObj,
  } = useIndustryAnalysis(deepest?.level ?? null, deepest?.id ?? null, fromDate, toDate);

  const breakdown = useAnalysisBreakdown(analysis);

  const nextLevelLabel =
    LEVEL_LABELS[LEVELS[Math.min(deepestLevelIndex + 1, MAX_LEVELS - 1)]];

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
        />

        <BackLink href="/" />

        {rangeInvalid && (
          <div className="bg-zinc-100 border border-zinc-300 text-zinc-700 text-xs font-bold px-4 py-3 rounded-xl">
            The start date must be before the end date.
          </div>
        )}

        <div>
          <h2 className="text-base md:text-lg font-black text-zinc-900 tracking-tight">
            Industry Analysis (SLSIC)
          </h2>
          <p className="text-xs text-zinc-400 mt-1 font-medium">
            Analyze through the industry classification hierarchy
          </p>
        </div>

        <ClassificationHierarchyPicker
          levels={levels}
          path={path}
          isInitializing={isInitializing}
          initializingLabel="Loading industry sectors..."
          initialLoadError={industrySectorsQuery.isError}
          initialLoadErrorLabel="Failed to load industry sectors"
          initialLoadErrorMessage={
            industrySectorsQuery.error instanceof Error
              ? industrySectorsQuery.error.message
              : undefined
          }
          emptyLabel="No industry sectors available."
          prunedNotice={prunedNotice}
          onSelect={handleSelect}
          onAddLevel={handleAddLevel}
          onRemoveLevel={handleRemoveLevel}
        />

        {deepest && (
          <div className="bg-white border border-zinc-200 rounded-xl shadow-sm">
            <AnalysisHeaderBar
              levelLabel={LEVEL_LABELS[deepest.level]}
              nodeName={deepest.name}
              nodeTotal={breakdown.nodeTotal}
              isLoading={analysisLoading}
            />
            <AnalysisBreakdownSection
              viewMode={viewMode}
              onViewModeChange={setViewMode}
              isLoading={analysisLoading}
              isError={analysisError}
              errorMessage={
                analysisErrorObj instanceof Error ? analysisErrorObj.message : undefined
              }
              accent={INDUSTRY_ACCENT}
              childrenBreakdownTitle={`${nextLevelLabel} Breakdown`}
              childrenBreakdownDescription={`Vacancy count of each ${nextLevelLabel.toLowerCase()} under ${deepest.name}`}
              childrenBreakdown={breakdown.childrenBreakdown}
              sectorData={breakdown.sectorData}
              expData={breakdown.expData}
              provinceData={breakdown.provinceData}
              eduData={breakdown.eduData}
              formalData={breakdown.formalData}
              genderData={breakdown.genderData}
              nvqData={breakdown.nvqData}
              remoteData={breakdown.remoteData}
              contractData={breakdown.contractData}
              remoteShare={breakdown.remoteShare}
              contractShare={breakdown.contractShare}
            />
          </div>
        )}
      </div>
    </div>
  );
}

export default function IndustryAnalysisPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-zinc-50 flex items-center justify-center text-xs text-zinc-400">
          Loading...
        </div>
      }
    >
      <IndustryAnalysis />
    </Suspense>
  );
}
