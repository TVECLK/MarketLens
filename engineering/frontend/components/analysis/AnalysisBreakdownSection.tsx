"use client";

import ViewModeToggle, { type AnalysisViewMode } from "@/components/analysis/ViewModeToggle";
import AnalysisBreakdownCharts from "@/components/analysis/AnalysisBreakdownCharts";
import AnalysisBreakdownTables from "@/components/analysis/AnalysisBreakdownTables";
import type { HierarchyAccent } from "@/lib/chart-theme";
import type { ChartDatum, ShareDatum } from "@/types/chart";

export type { AnalysisViewMode } from "@/components/analysis/ViewModeToggle";

interface AnalysisBreakdownSectionProps {
  viewMode: AnalysisViewMode;
  onViewModeChange: (mode: AnalysisViewMode) => void;
  isLoading: boolean;
  isError: boolean;
  errorMessage?: string;
  accent: HierarchyAccent;
  childrenBreakdownTitle: string;
  childrenBreakdownDescription: string;
  childrenBreakdown: ChartDatum[];
  sectorData: ChartDatum[];
  expData: ChartDatum[];
  provinceData: ChartDatum[];
  eduData: ChartDatum[];
  formalData: ChartDatum[];
  genderData: ChartDatum[];
  nvqData: ChartDatum[];
  remoteData: ChartDatum[];
  contractData: ChartDatum[];
  remoteShare: ShareDatum[];
  contractShare: ShareDatum[];
}

export default function AnalysisBreakdownSection({
  viewMode,
  onViewModeChange,
  isLoading,
  isError,
  errorMessage,
  accent,
  childrenBreakdownTitle,
  childrenBreakdownDescription,
  childrenBreakdown,
  sectorData,
  expData,
  provinceData,
  eduData,
  formalData,
  genderData,
  nvqData,
  remoteData,
  contractData,
  remoteShare,
  contractShare,
}: AnalysisBreakdownSectionProps) {
  return (
    <>
      <ViewModeToggle viewMode={viewMode} onChange={onViewModeChange} />

      {isLoading ? (
        <div className="p-10 text-center text-xs text-zinc-400">Loading analytics...</div>
      ) : isError ? (
        <div className="p-10 text-center text-xs text-red-500 font-bold">
          Failed to load analytics
          {errorMessage ? `: ${errorMessage}` : "."}
        </div>
      ) : viewMode === "chart" ? (
        <AnalysisBreakdownCharts
          accent={accent}
          childrenBreakdown={childrenBreakdown}
          childrenBreakdownTitle={childrenBreakdownTitle}
          childrenBreakdownDescription={childrenBreakdownDescription}
          sectorData={sectorData}
          expData={expData}
          provinceData={provinceData}
          eduData={eduData}
          formalData={formalData}
          genderData={genderData}
          nvqData={nvqData}
          remoteShare={remoteShare}
          contractShare={contractShare}
        />
      ) : (
        <AnalysisBreakdownTables
          childrenBreakdown={childrenBreakdown}
          childrenBreakdownTitle={childrenBreakdownTitle}
          sectorData={sectorData}
          expData={expData}
          provinceData={provinceData}
          eduData={eduData}
          formalData={formalData}
          genderData={genderData}
          nvqData={nvqData}
          remoteData={remoteData}
          contractData={contractData}
        />
      )}
    </>
  );
}
