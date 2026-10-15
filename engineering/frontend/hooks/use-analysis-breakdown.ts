"use client";

import { useMemo } from "react";
import type { AnalysisResponse } from "@/types/analysis";
import type { ChartDatum, ShareDatum } from "@/types/chart";

interface AnalysisBreakdown {
  nodeTotal: number;
  childrenBreakdown: ChartDatum[];
  sectorData: ChartDatum[];
  expData: ChartDatum[];
  eduData: ChartDatum[];
  formalData: ChartDatum[];
  genderData: ChartDatum[];
  nvqData: ChartDatum[];
  provinceData: ChartDatum[];
  contractData: ChartDatum[];
  remoteData: ChartDatum[];
  remoteShare: ShareDatum[];
  contractShare: ShareDatum[];
}

function toShareBars(rows: ChartDatum[]): ShareDatum[] {
  const total = rows.reduce((sum, r) => sum + r.value, 0) || 1;
  return rows.map((r) => ({
    label: r.name,
    share: Math.round((r.value / total) * 100),
  }));
}

/**
 * Shapes a classification-hierarchy analysis response (identical structure
 * for industry and occupation, see `types/analysis.ts`) into the flat
 * {name, value} arrays the breakdown charts and tables consume.
 */
export function useAnalysisBreakdown(
  analysis: AnalysisResponse | undefined,
): AnalysisBreakdown {
  const nodeTotal = analysis?.total_job_count.total_job_count ?? 0;

  const childrenBreakdown = useMemo(() => {
    if (!analysis?.children) return [];
    return analysis.children.children
      .map((c) => ({ name: c.name, value: c.open_job_count }))
      .sort((a, b) => b.value - a.value);
  }, [analysis]);

  const sectorData = useMemo(
    () =>
      (analysis?.employment_sector.employment_sectors ?? []).map((s) => ({
        name: s.sector,
        value: s.open_job_count,
      })),
    [analysis],
  );
  const expData = useMemo(
    () =>
      (analysis?.experience.experiences ?? []).map((e) => ({
        name: e.name,
        value: e.open_job_count,
      })),
    [analysis],
  );
  const eduData = useMemo(
    () =>
      (analysis?.education.education_levels ?? []).map((e) => ({
        name: e.level,
        value: e.open_job_count,
      })),
    [analysis],
  );
  const formalData = useMemo(
    () =>
      (analysis?.formality.formalities ?? []).map((f) => ({
        name: f.formality_type,
        value: f.open_job_count,
      })),
    [analysis],
  );
  const genderData = useMemo(
    () =>
      (analysis?.gender.genders ?? []).map((g) => ({
        name: g.gender_type,
        value: g.open_job_count,
      })),
    [analysis],
  );
  const nvqData = useMemo(
    () =>
      (analysis?.vocational_education.vocational_educations ?? []).map((v) => ({
        name: v.level,
        value: v.open_job_count,
      })),
    [analysis],
  );
  const provinceData = useMemo(
    () =>
      (analysis?.province.provinces ?? []).map((p) => ({
        name: p.province,
        value: p.open_job_count,
      })),
    [analysis],
  );
  const contractData = useMemo(
    () =>
      (analysis?.job_type.job_types ?? []).map((j) => ({
        name: j.type,
        value: j.open_job_count,
      })),
    [analysis],
  );
  const remoteData = useMemo(() => {
    const r = analysis?.remote_onsite_hybrid.remote_vs_onsite_vs_hybrid;
    if (!r) return [];
    return [
      { name: "On-Site", value: r.on_site_count },
      { name: "Remote", value: r.remote_count },
      { name: "Hybrid", value: r.hybrid_count },
    ];
  }, [analysis]);

  const remoteShare = useMemo(() => toShareBars(remoteData), [remoteData]);
  const contractShare = useMemo(() => toShareBars(contractData), [contractData]);

  return {
    nodeTotal,
    childrenBreakdown,
    sectorData,
    expData,
    eduData,
    formalData,
    genderData,
    nvqData,
    provinceData,
    contractData,
    remoteData,
    remoteShare,
    contractShare,
  };
}
