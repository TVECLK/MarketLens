import type { HierarchyRow } from "@/types/kpi";
import type { ActionPermissions } from "@/lib/action-permissions";

export interface MutationLike<TVars> {
  mutate: (vars: TVars, opts?: { onSuccess?: () => void; onError?: (err: unknown) => void }) => void;
  isPending: boolean;
}

export interface FlatItem {
  id: number;
  label: string;
}

export interface FlatBundle {
  isLoading: boolean;
  isError: boolean;
  items: FlatItem[];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see note above
  createMut: MutationLike<any>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see note above
  updateMut: MutationLike<{ id: number; payload: any }>;
  deleteMut: MutationLike<number>;
  labelField: string;
  permissions: ActionPermissions;
}

/** The top level of a hierarchy, loaded as a plain list. */
export interface HierarchyRoots {
  isLoading: boolean;
  isError: boolean;
  items: HierarchyRow[];
}

/** The mutations and permissions of one hierarchy level. */
export interface HierarchyLevelMutations {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see note above
  createMut: MutationLike<any>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see note above
  updateMut: MutationLike<{ id: number; payload: any }>;
  deleteMut: MutationLike<number>;
  parentField: string | null; // e.g. "major_group_id", null for the top level
  permissions: ActionPermissions;
}

/** The tree item shown in the detail panel, with the items above it (top -> parent). */
export interface HierarchySelection {
  levelIndex: number;
  row: HierarchyRow;
  ancestors: HierarchySelection[];
}
