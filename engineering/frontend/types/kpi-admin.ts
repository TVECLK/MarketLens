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

export interface HierarchyLevelBundle {
  isLoading: boolean;
  isError: boolean;
  items: HierarchyRow[];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see note above
  createMut: MutationLike<any>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- see note above
  updateMut: MutationLike<{ id: number; payload: any }>;
  deleteMut: MutationLike<number>;
  parentField: string | null; // e.g. "major_group_id", null for the top level
  total?: number;
  page?: number;
  setPage?: (p: number) => void;
}
