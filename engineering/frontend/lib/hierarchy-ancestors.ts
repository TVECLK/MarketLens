import type { HierarchyRow } from "@/types/kpi";
import type { HierarchyRoots, HierarchySelection } from "@/types/kpi-admin";

/**
 * The chain from the top level down to `parent`, as a selection's `ancestors`
 * (top -> parent). Each level's own parent is looked up by id in `lists`, where
 * `lists[i]` holds every item of level i. The chain stops early if an item's
 * parent isn't in its list (for example while that list is still loading).
 */
export function buildAncestors(
  lists: readonly HierarchyRoots[],
  parentLevelIndex: number,
  parent: HierarchyRow,
): HierarchySelection[] {
  const chain: { levelIndex: number; row: HierarchyRow }[] = [];
  let levelIndex = parentLevelIndex;
  let row: HierarchyRow | undefined = parent;

  while (row) {
    chain.unshift({ levelIndex, row });
    if (levelIndex === 0 || row.parentId == null) break;
    const parentId: number = row.parentId;
    row = lists[levelIndex - 1]?.items.find((item) => item.id === parentId);
    levelIndex -= 1;
  }

  return chain.reduce<HierarchySelection[]>((acc, node) => [...acc, { ...node, ancestors: [...acc] }], []);
}
