import type { HierarchyRow } from "@/types/kpi";

// Adapts a top-level hierarchy list to the common row shape (HierarchyRow).

export function normTop(
  list: { id: number; name: string; code: string }[] | undefined,
): HierarchyRow[] {
  return (list ?? []).map((i) => ({
    id: i.id,
    name: i.name,
    code: i.code,
    parentId: null,
    parentLabel: null,
  }));
}
