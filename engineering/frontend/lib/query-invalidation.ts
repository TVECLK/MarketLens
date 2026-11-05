import type { QueryClient } from "@tanstack/react-query";
import { OCCUPATION_QUERY_KEYS, INDUSTRY_QUERY_KEYS } from "@/lib/kpi-config";

/**
 * Roots of the query keys the dashboard and analysis pages use. They are
 * separate from the admin lists (`[key, "list"]`), so editing reference data
 * in /kpis has to invalidate them explicitly or the pages keep showing the
 * cached values until `staleTime` runs out.
 */
export const DASHBOARD_QUERY_ROOTS = [
  "occupation",
  "industry",
  "analysis",
  "dashboard",
  "occupation-skills",
] as const;

/**
 * The admin lists to refresh when `key`'s list changes. A hierarchy level's
 * rows show their parent's name, so a change refreshes every level of that
 * hierarchy; a flat KPI only refreshes its own list.
 */
export function relatedListKeys(key: string): string[] {
  const occupationKeys: string[] = OCCUPATION_QUERY_KEYS;
  const industryKeys: string[] = INDUSTRY_QUERY_KEYS;
  if (occupationKeys.includes(key)) return occupationKeys;
  if (industryKeys.includes(key)) return industryKeys;
  return [key];
}

/** Marks the given admin lists and every dashboard/analysis query stale. */
export function invalidateAfterKpiChange(queryClient: QueryClient, listKeys: readonly string[]) {
  for (const key of listKeys) {
    queryClient.invalidateQueries({ queryKey: [key, "list"] });
  }
  for (const root of DASHBOARD_QUERY_ROOTS) {
    queryClient.invalidateQueries({ queryKey: [root] });
  }
}
