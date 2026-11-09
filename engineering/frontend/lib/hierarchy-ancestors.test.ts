import { describe, expect, it } from "vitest";
import { buildAncestors } from "./hierarchy-ancestors";
import type { HierarchyRoots } from "@/types/kpi-admin";

const row = (id: number, parentId: number | null) => ({ id, name: `n${id}`, code: `c${id}`, parentId, parentLabel: null });
const list = (...items: ReturnType<typeof row>[]): HierarchyRoots => ({ isLoading: false, isError: false, items });

describe("buildAncestors", () => {
  const lists = [list(row(1, null), row(2, null)), list(row(10, 1), row(20, 2)), list(row(100, 10))];

  it("returns just the parent for a top-level parent", () => {
    const chain = buildAncestors(lists, 0, row(2, null));
    expect(chain.map((n) => [n.levelIndex, n.row.id])).toEqual([[0, 2]]);
  });

  it("walks up to the top level, top first", () => {
    const chain = buildAncestors(lists, 2, row(100, 10));
    expect(chain.map((n) => [n.levelIndex, n.row.id])).toEqual([[0, 1], [1, 10], [2, 100]]);
  });

  it("gives each node the nodes above it as its ancestors", () => {
    const chain = buildAncestors(lists, 2, row(100, 10));
    expect(chain[2].ancestors.map((a) => a.row.id)).toEqual([1, 10]);
    expect(chain[0].ancestors).toEqual([]);
  });

  it("stops early when an ancestor isn't in its list", () => {
    const chain = buildAncestors([list(), list()], 1, row(10, 1));
    expect(chain.map((n) => n.row.id)).toEqual([10]);
  });
});
