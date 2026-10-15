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
import SkillAnalysisPanel from "@/components/occupations/SkillAnalysisPanel";
import { useScrolled } from "@/hooks/use-scrolled";
import { useUrlAwareDateRange } from "@/hooks/use-url-aware-date-range";
import {
  useHierarchyPath,
  useHierarchyDrilldown,
  type HierarchyLevelState,
} from "@/hooks/use-hierarchy-drilldown";
import { useAnalysisBreakdown } from "@/hooks/use-analysis-breakdown";
import {
  useMajorGroup,
  useSubMajorGroups,
  useMinorGroups,
  useUnitGroups,
  useOccupationGroups,
  useOccupationAnalysis,
  useOccupationSkills,
} from "@/hooks/use-occupation";
import { CHART_COLORS } from "@/lib/chart-theme";
import { getDataRangeStart, getDataRangeEnd } from "@/lib/date-range";

const DATE_RANGE_STORAGE_KEY = "lmis-date-range:v2";

type OccupationLevel =
  | "major-group"
  | "sub-major-group"
  | "minor-group"
  | "unit-group"
  | "occupation-group";

const LEVELS: OccupationLevel[] = [
  "major-group",
  "sub-major-group",
  "minor-group",
  "unit-group",
  "occupation-group",
];
const LEVEL_LABELS: Record<OccupationLevel, string> = {
  "major-group": "Major Group",
  "sub-major-group": "Sub Major Group",
  "minor-group": "Minor Group",
  "unit-group": "Unit Group",
  "occupation-group": "Occupation Group",
};
const MAX_LEVELS = LEVELS.length;

const OCCUPATION_ACCENT = { primary: CHART_COLORS.indigo, secondary: CHART_COLORS.teal };

function OccupationAnalysis() {
  const { fromDate, toDate, setFromDate, setToDate, rangeInvalid } =
    useUrlAwareDateRange(DATE_RANGE_STORAGE_KEY);
  const scrolled = useScrolled();
  const [activeTab, setActiveTab] = useState<"analytics" | "skills">("analytics");
  const [viewMode, setViewMode] = useState<AnalysisViewMode>("chart");
  const dataRangeEnd = getDataRangeEnd();
  const dataRangeStart = getDataRangeStart();

  const [path, setPath] = useHierarchyPath<OccupationLevel>();

  const majorGroupQuery = useMajorGroup(fromDate, toDate);
  const majorGroups = useMemo(() => majorGroupQuery.data?.major_groups ?? [], [majorGroupQuery.data]);

  const subMajorGroupsQuery = useSubMajorGroups(path[0]?.id ?? null, fromDate, toDate);
  const minorGroupsQuery = useMinorGroups(path[1]?.id ?? null, fromDate, toDate);
  const unitGroupsQuery = useUnitGroups(path[2]?.id ?? null, fromDate, toDate);
  const occupationGroupsQuery = useOccupationGroups(path[3]?.id ?? null, fromDate, toDate);

  const levels: HierarchyLevelState<OccupationLevel>[] = [
    {
      key: "major-group",
      label: LEVEL_LABELS["major-group"],
      options: majorGroups,
      isLoading: majorGroupQuery.isLoading,
      isSuccess: majorGroupQuery.isSuccess,
    },
    {
      key: "sub-major-group",
      label: LEVEL_LABELS["sub-major-group"],
      options: subMajorGroupsQuery.data?.sub_major_groups ?? [],
      isLoading: subMajorGroupsQuery.isLoading,
      isSuccess: subMajorGroupsQuery.isSuccess,
    },
    {
      key: "minor-group",
      label: LEVEL_LABELS["minor-group"],
      options: minorGroupsQuery.data?.minor_groups ?? [],
      isLoading: minorGroupsQuery.isLoading,
      isSuccess: minorGroupsQuery.isSuccess,
    },
    {
      key: "unit-group",
      label: LEVEL_LABELS["unit-group"],
      options: unitGroupsQuery.data?.unit_groups ?? [],
      isLoading: unitGroupsQuery.isLoading,
      isSuccess: unitGroupsQuery.isSuccess,
    },
    {
      key: "occupation-group",
      label: LEVEL_LABELS["occupation-group"],
      options: occupationGroupsQuery.data?.occupation_groups ?? [],
      isLoading: occupationGroupsQuery.isLoading,
      isSuccess: occupationGroupsQuery.isSuccess,
    },
  ];

  const { prunedNotice, handleSelect, handleAddLevel, handleRemoveLevel } =
    useHierarchyDrilldown(path, setPath, levels);

  const deepest = path[path.length - 1] ?? null;
  const deepestLevelIndex = path.length - 1;
  const isInitializing = majorGroupQuery.isLoading;

  const {
    data: analysis,
    isLoading: analysisLoading,
    isError: analysisError,
    error: analysisErrorObj,
  } = useOccupationAnalysis(deepest?.level ?? null, deepest?.id ?? null, fromDate, toDate);

  const breakdown = useAnalysisBreakdown(analysis);

  const {
    data: skillsData,
    isLoading: skillsLoading,
    isError: skillsError,
  } = useOccupationSkills(
    deepest?.level ?? null,
    deepest?.id ?? null,
    fromDate,
    toDate,
    activeTab === "skills",
  );

  const topSkills = useMemo(
    () =>
      (skillsData?.top_15_skills.skills ?? []).map((s) => ({
        name: s.skill,
        value: s.open_job_count,
      })),
    [skillsData],
  );
  const allSkills = useMemo(() => skillsData?.all_skills.skills ?? [], [skillsData]);
  const totalSkillMentions = useMemo(
    () => allSkills.reduce((sum, s) => sum + s.open_job_count, 0) || 1,
    [allSkills],
  );
  const topEmployers = useMemo(() => skillsData?.top_hiring_employers.employers ?? [], [skillsData]);

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
            Occupation Analysis (SLSO)
          </h2>
          <p className="text-xs text-zinc-400 mt-1 font-medium">
            Analyze through the occupation classification hierarchy
          </p>
        </div>

        <ClassificationHierarchyPicker
          levels={levels}
          path={path}
          isInitializing={isInitializing}
          initializingLabel="Loading occupation groups..."
          initialLoadError={majorGroupQuery.isError}
          initialLoadErrorLabel="Failed to load occupation groups"
          initialLoadErrorMessage={
            majorGroupQuery.error instanceof Error ? majorGroupQuery.error.message : undefined
          }
          emptyLabel="No occupation groups available."
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

            <div className="px-5 border-b border-zinc-200 flex gap-6">
              <button
                onClick={() => setActiveTab("analytics")}
                className={`cursor-pointer py-3 text-xs font-bold border-b-2 -mb-px transition-colors ${activeTab === "analytics" ? "border-zinc-900 text-zinc-900" : "border-transparent text-zinc-400 hover:text-zinc-600"}`}
              >
                Analytics
              </button>
              <button
                onClick={() => setActiveTab("skills")}
                className={`cursor-pointer py-3 text-xs font-bold border-b-2 -mb-px transition-colors ${activeTab === "skills" ? "border-zinc-900 text-zinc-900" : "border-transparent text-zinc-400 hover:text-zinc-600"}`}
              >
                Skill Analysis
              </button>
            </div>

            {activeTab === "analytics" && (
              <AnalysisBreakdownSection
                viewMode={viewMode}
                onViewModeChange={setViewMode}
                isLoading={analysisLoading}
                isError={analysisError}
                errorMessage={
                  analysisErrorObj instanceof Error ? analysisErrorObj.message : undefined
                }
                accent={OCCUPATION_ACCENT}
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
            )}

            {activeTab === "skills" && (
              <SkillAnalysisPanel
                isLoading={skillsLoading}
                isError={skillsError}
                topSkills={topSkills}
                allSkills={allSkills}
                totalSkillsCount={skillsData?.all_skills.total ?? 0}
                totalSkillMentions={totalSkillMentions}
                topEmployers={topEmployers}
              />
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default function OccupationAnalysisPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-zinc-50 flex items-center justify-center text-xs text-zinc-400">
          Loading...
        </div>
      }
    >
      <OccupationAnalysis />
    </Suspense>
  );
}
