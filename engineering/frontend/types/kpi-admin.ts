import type { HierarchyRow } from "@/types/kpi";
import type { HierarchyId } from "@/lib/kpi-config";

// Deliberately loosely typed (payload: any) — six flat KPIs and ten hierarchy
// levels each have a different payload shape server-side, and collapsing
// them into one generic table/form means the container has to erase those
// specific types. Each concrete hook still enforces its own payload type at
// the call site (useCreateEmploymentSector, etc.) — only this bundle
// wiring loses that specificity, by design.
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

export type SectionKey = "flat" | HierarchyId;
