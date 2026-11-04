import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
import { fetchJson } from "@/lib/fetch-json";
import type { AnalysisResponse } from "@/types/analysis";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;

export type HierarchyStandard = "occupation" | "industry";

interface HierarchyRequest {
  queryKey: readonly unknown[];
  url: string;
  enabled: boolean;
}

function dateQuery(fromDate: string, toDate: string): string {
  return new URLSearchParams({ "from-date": fromDate, "to-date": toDate }).toString();
}

/** The top level of a hierarchy: no parent. */
export function hierarchyRootRequest(
  standard: HierarchyStandard,
  path: string,
  fromDate: string,
  toDate: string,
): HierarchyRequest {
  return {
    queryKey: [standard, path, fromDate, toDate],
    url: `${BASE}/${standard}/${path}?${dateQuery(fromDate, toDate)}`,
    enabled: Boolean(fromDate) && Boolean(toDate),
  };
}

/** The children of one parent, e.g. the sub major groups of a major group. */
export function hierarchyChildrenRequest(
  standard: HierarchyStandard,
  parentPath: string,
  childPath: string,
  parentId: number | null,
  fromDate: string,
  toDate: string,
): HierarchyRequest {
  return {
    queryKey: [standard, childPath, parentId, fromDate, toDate],
    url: `${BASE}/${standard}/${parentPath}/${parentId}/${childPath}?${dateQuery(fromDate, toDate)}`,
    enabled: parentId !== null && Boolean(fromDate) && Boolean(toDate),
  };
}

/** The analysis of one item at one level; the occupation and industry responses share a shape. */
export function hierarchyAnalysisRequest(
  standard: HierarchyStandard,
  level: string | null,
  id: number | null,
  fromDate: string,
  toDate: string,
): HierarchyRequest {
  return {
    queryKey: ["analysis", standard, level, id, fromDate, toDate],
    url: `${BASE}/analysis/${standard}/${level}/${id}?${dateQuery(fromDate, toDate)}`,
    enabled: level !== null && id !== null && Boolean(fromDate) && Boolean(toDate) && fromDate <= toDate,
  };
}

export function useHierarchyRoot<T>(standard: HierarchyStandard, path: string, fromDate: string, toDate: string) {
  const { queryKey, url, enabled } = hierarchyRootRequest(standard, path, fromDate, toDate);
  return useQuery<T>({
    queryKey,
    queryFn: ({ signal }) => fetchJson<T>(url, { signal }),
    enabled,
  });
}

export function useHierarchyChildren<T>(
  standard: HierarchyStandard,
  parentPath: string,
  childPath: string,
  parentId: number | null,
  fromDate: string,
  toDate: string,
) {
  const { queryKey, url, enabled } = hierarchyChildrenRequest(standard, parentPath, childPath, parentId, fromDate, toDate);
  return useQuery<T>({
    queryKey,
    queryFn: ({ signal }) => fetchJson<T>(url, { signal }),
    enabled,
  });
}

export function useHierarchyAnalysis(
  standard: HierarchyStandard,
  level: string | null,
  id: number | null,
  fromDate: string,
  toDate: string,
) {
  const { queryKey, url, enabled } = hierarchyAnalysisRequest(standard, level, id, fromDate, toDate);
  return useQuery<AnalysisResponse>({
    queryKey,
    queryFn: ({ signal }) =>
      fetchJson<AnalysisResponse>(url, { signal, errorMessage: `Failed to fetch ${standard} analysis` }),
    enabled,
  });
}
