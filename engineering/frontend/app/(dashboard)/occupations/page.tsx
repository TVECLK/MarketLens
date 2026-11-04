"use client";

import { useMemo } from "react";
import BackLink from "@/components/shared/BackLink";
import ClassificationHierarchyPicker from "@/components/analysis/ClassificationHierarchyPicker";
import AnalysisHeaderBar from "@/components/analysis/AnalysisHeaderBar";
import AnalysisBreakdownSection from "@/components/analysis/AnalysisBreakdownSection";
import SkillAnalysisPanel from "@/components/occupations/SkillAnalysisPanel";
import { useDateRange } from "@/providers/date-range-provider";
import { useAnalysisViewMode } from "@/hooks/use-analysis-view-mode";
import { useUrlEnumParam } from "@/hooks/use-url-enum-param";
import { URL_PARAMS } from "@/lib/url-params";
import {
  useHierarchyPath,
  selectedIdAt,
  useHierarchyDrilldown,
  type HierarchyLevelState,
} from "@/hooks/use-hierarchy-drilldown";
import { useAnalysisBreakdown } from "@/hooks/use-analysis-breakdown";
import { useHierarchyAnalysis } from "@/hooks/use-hierarchy-queries";
import {
  useMajorGroup,
  useSubMajorGroups,
  useMinorGroups,
  useUnitGroups,
  useOccupationGroups,
  useOccupationSkills,
} from "@/hooks/use-occupation";
import { CHART_COLORS } from "@/lib/chart-theme";
import {
  OCCUPATION_LEVELS as LEVELS,
  OCCUPATION_LEVEL_LABELS as LEVEL_LABELS,
  type OccupationLevel,
} from "@/lib/kpi-config";

const MAX_LEVELS = LEVELS.length;

const OCCUPATION_TABS = ["analytics", "skills"] as const;

const OCCUPATION_ACCENT = { primary: CHART_COLORS.indigo, secondary: CHART_COLORS.teal };

export default function OccupationAnalysisPage() {
  const { fromDate, toDate } = useDateRange();
  const [activeTab, setActiveTab] = useUrlEnumParam(URL_PARAMS.tab, OCCUPATION_TABS, "analytics");
  const [viewMode, setViewMode] = useAnalysisViewMode();

  const [ids, setIds] = useHierarchyPath(LEVELS);

  const majorGroupQuery = useMajorGroup(fromDate, toDate);
  const majorGroups = useMemo(() => majorGroupQuery.data?.major_groups ?? [], [majorGroupQuery.data]);

  const subMajorGroupsQuery = useSubMajorGroups(selectedIdAt(ids, LEVELS, "major-group"), fromDate, toDate);
  const minorGroupsQuery = useMinorGroups(selectedIdAt(ids, LEVELS, "sub-major-group"), fromDate, toDate);
  const unitGroupsQuery = useUnitGroups(selectedIdAt(ids, LEVELS, "minor-group"), fromDate, toDate);
  const occupationGroupsQuery = useOccupationGroups(selectedIdAt(ids, LEVELS, "unit-group"), fromDate, toDate);

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

  const { path, prunedNotice, handleSelect, handleAddLevel, handleRemoveLevel } =
    useHierarchyDrilldown(ids, setIds, levels, `${fromDate}|${toDate}`);

  const deepest = path[path.length - 1] ?? null;
  const deepestLevelIndex = path.length - 1;
  const isInitializing = majorGroupQuery.isLoading;

  const {
    data: analysis,
    isLoading: analysisLoading,
    isError: analysisError,
    error: analysisErrorObj,
  } = useHierarchyAnalysis("occupation", deepest?.level ?? null, deepest?.id ?? null, fromDate, toDate);

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
      (skillsData?.top_15_skills?.skills ?? []).map((s) => ({
        name: s.skill,
        value: s.open_job_count,
      })),
    [skillsData],
  );
  const allSkills = useMemo(() => skillsData?.all_skills?.skills ?? [], [skillsData]);
  const totalSkillMentions = useMemo(
    () => allSkills.reduce((sum, s) => sum + s.open_job_count, 0) || 1,
    [allSkills],
  );
  const topEmployers = useMemo(() => skillsData?.top_hiring_employers?.employers ?? [], [skillsData]);

  const nextLevelLabel =
    LEVEL_LABELS[LEVELS[Math.min(deepestLevelIndex + 1, MAX_LEVELS - 1)]];

  return (
    <div className="px-4 md:px-8 py-2 md:py-4 space-y-6 w-full pb-20">
      <BackLink href="/" />

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
              isEmpty={breakdown.nodeTotal === 0}
              sectionErrors={breakdown.sectionErrors}
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
              totalSkillsCount={skillsData?.all_skills?.total ?? 0}
              totalSkillMentions={totalSkillMentions}
              topEmployers={topEmployers}
              sectionErrors={skillsData?.errors ?? {}}
            />
          )}
        </div>
      )}
    </div>
  );
}
