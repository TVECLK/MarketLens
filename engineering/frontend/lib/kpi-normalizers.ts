import type { HierarchyRow } from "@/types/kpi";

// Each raw hierarchy API type has a different label/parent field name;
// these adapt every level down to one common shape (HierarchyRow) so the
// table/form JSX only needs to be written once instead of ten times.

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

export function normChild(
  list: unknown[] | undefined,
  parentIdField: string,
  parentField: string,
): HierarchyRow[] {
  return (list ?? []).map((raw) => {
    const item = raw as Record<string, unknown>;
    const parentId = item[parentIdField] as number;
    const parent = item[parentField] as { name: string; code: string } | null;
    return {
      id: item.id as number,
      name: item.name as string,
      code: item.code as string,
      parentId,
      parentLabel: parent ? `${parent.code} — ${parent.name}` : null,
    };
  });
}
