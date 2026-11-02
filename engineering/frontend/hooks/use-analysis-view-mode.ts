"use client";

import { useUrlEnumParam } from "@/hooks/use-url-enum-param";
import { ANALYSIS_VIEW_MODES, type AnalysisViewMode } from "@/components/analysis/ViewModeToggle";
import { URL_PARAMS } from "@/lib/url-params";

/** Chart/table toggle shared by the industry and occupation analysis pages. */
export function useAnalysisViewMode(): [AnalysisViewMode, (mode: AnalysisViewMode) => void] {
  return useUrlEnumParam(URL_PARAMS.view, ANALYSIS_VIEW_MODES, "chart");
}
