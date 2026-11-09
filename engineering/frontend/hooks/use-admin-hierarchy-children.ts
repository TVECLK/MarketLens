import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
import { fetchJson } from "@/lib/fetch-json";
import { HIERARCHIES, type HierarchyId } from "@/lib/kpi-config";
import type { HierarchyRow } from "@/types/kpi";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;

/**
 * The by-parent request for the children of one item, e.g. the sub major
 * groups of a major group. No date range is sent, so the backend returns
 * only the currently active rows. The query key starts with the hierarchy id,
 * so every create/update/delete (which invalidates the "occupation" /
 * "industry" roots) refreshes the open branches too.
 */
export function adminChildrenRequest(hierarchyId: HierarchyId, parentLevelIndex: number, parentId: number) {
  const levels = HIERARCHIES[hierarchyId].levels;
  const parent = levels[parentLevelIndex];
  const child = levels[parentLevelIndex + 1];
  return {
    queryKey: [hierarchyId, "tree-children", child.queryKey, parentId] as const,
    url: `${BASE}/${hierarchyId}/${parent.queryKey}/${parentId}/${child.queryKey}`,
    // The response lists the rows under the child level's name, e.g. "sub_major_groups".
    listField: child.queryKey.replaceAll("-", "_"),
  };
}

/** Lazily loads the direct children of one item. Only call it for a non-leaf level. */
export function useAdminHierarchyChildren(hierarchyId: HierarchyId, parentLevelIndex: number, parentId: number) {
  const { queryKey, url, listField } = adminChildrenRequest(hierarchyId, parentLevelIndex, parentId);
  return useQuery<HierarchyRow[]>({
    queryKey,
    queryFn: async ({ signal }) => {
      const data = await fetchJson<Record<string, { id: number; name: string; code: string }[] | null>>(url, {
        signal,
        errorMessage: "Failed to load child items",
      });
      return (data[listField] ?? []).map((item) => ({
        id: item.id,
        name: item.name,
        code: item.code,
        parentId,
        parentLabel: null,
      }));
    },
  });
}
