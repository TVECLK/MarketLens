"use client";

import { useMemo } from "react";
import BackLink from "@/components/shared/BackLink";
import ClassificationHierarchyPicker from "@/components/analysis/ClassificationHierarchyPicker";
import AnalysisHeaderBar from "@/components/analysis/AnalysisHeaderBar";
import AnalysisBreakdownSection from "@/components/analysis/AnalysisBreakdownSection";
import { useDateRange } from "@/providers/date-range-provider";
import { useAnalysisViewMode } from "@/hooks/use-analysis-view-mode";
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
import {
  INDUSTRY_LEVELS as LEVELS,
  INDUSTRY_LEVEL_LABELS as LEVEL_LABELS,
  type IndustryLevel,
} from "@/lib/kpi-config";

const MAX_LEVELS = LEVELS.length;

const INDUSTRY_ACCENT = { primary: CHART_COLORS.teal, secondary: CHART_COLORS.indigo };

export default function IndustryAnalysisPage() {
  const { fromDate, toDate } = useDateRange();
  const [viewMode, setViewMode] = useAnalysisViewMode();

  const [ids, setIds] = useHierarchyPath(LEVELS);

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

  const industryDivisionsQuery = useIndustryDivisions(ids[0] ?? null, fromDate, toDate);
  const industryGroupsQuery = useIndustryGroups(ids[1] ?? null, fromDate, toDate);
  const industryClassesQuery = useIndustryClasses(ids[2] ?? null, fromDate, toDate);
  const industrySubclassesQuery = useIndustrySubclasses(ids[3] ?? null, fromDate, toDate);

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

  const { path, prunedNotice, handleSelect, handleAddLevel, handleRemoveLevel } =
    useHierarchyDrilldown(ids, setIds, levels, `${fromDate}|${toDate}`);

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
    <div className="px-4 md:px-8 py-2 md:py-4 space-y-6 w-full pb-20">
      <BackLink href="/" />

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
            isEmpty={breakdown.nodeTotal === 0}
            sectionErrors={breakdown.sectionErrors}
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
  );
}
