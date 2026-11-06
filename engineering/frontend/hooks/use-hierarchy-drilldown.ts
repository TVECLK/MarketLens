"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { parseHierarchyIds, serializeHierarchyIds } from "@/lib/url-params";
import { useUpdateSearchParams } from "@/hooks/use-update-search-params";
import type { HierarchyNode } from "@/types/analysis";

export interface HierarchyLevelState<TLevelKey extends string> {
  key: TLevelKey;
  label: string;
  options: HierarchyNode[];
  isLoading: boolean;
  isSuccess: boolean;
}

export interface PathNode<TLevelKey extends string> extends HierarchyNode {
  level: TLevelKey;
}

/**
 * The id selected at `level`, or null when that level isn't selected yet (or
 * isn't one of `levels`). Looks the level up by name instead of by position,
 * so a query's parent id keeps pointing at the right level if `levels` is
 * ever reordered or shortened.
 */
export function selectedIdAt<TLevel extends string>(
  ids: readonly number[],
  levels: readonly TLevel[],
  level: TLevel,
): number | null {
  const index = levels.indexOf(level);
  return index >= 0 && index < ids.length ? ids[index] : null;
}

/**
 * Owns the drilldown selection, stored in the URL as one param per level
 * (`?major-group=2&sub-major-group=11`) so a drilldown can be shared as a
 * link. Returns the selected ids in level order. Call this first, before any
 * per-level react-query hooks — each level's query needs `ids[i - 1]` as an
 * argument, so the selection has to exist before those queries run.
 */
export function useHierarchyPath(
  levelKeys: readonly string[],
): [number[], (ids: number[]) => void] {
  const searchParams = useSearchParams();
  const updateSearchParams = useUpdateSearchParams();

  const ids = useMemo(
    () => parseHierarchyIds(searchParams, levelKeys),
    [searchParams, levelKeys],
  );
  const setIds = useCallback(
    (next: number[]) => updateSearchParams(serializeHierarchyIds(next, levelKeys)),
    [updateSearchParams, levelKeys],
  );

  return [ids, setIds];
}

interface UseHierarchyDrilldownResult<TLevelKey extends string> {
  path: PathNode<TLevelKey>[];
  prunedNotice: string | null;
  handleSelect: (levelIndex: number, idStr: string) => void;
  handleAddLevel: () => void;
  handleRemoveLevel: () => void;
}

/**
 * Drives a "classification hierarchy" drill-down: a chain of selects where
 * each level's options depend on the previous level's selection (e.g.
 * industry sector -> division -> group -> class -> subclass).
 *
 * Call `useHierarchyPath` first to get `ids`/`setIds`, call each level's
 * own react-query hook (Rules of Hooks requires a fixed, unconditional hook
 * per level — this can't be a loop), assemble the results into `levels`,
 * then call this hook with all three. `path` is `ids` resolved into nodes as
 * each level's options load.
 */
export function useHierarchyDrilldown<TLevelKey extends string>(
  ids: number[],
  setIds: (ids: number[]) => void,
  levels: HierarchyLevelState<TLevelKey>[],
  resetKey: string,
): UseHierarchyDrilldownResult<TLevelKey> {
  const [prunedNotice, setPrunedNotice] = useState<string | null>(null);

  const [noticeResetKey, setNoticeResetKey] = useState(resetKey);
  if (noticeResetKey !== resetKey) {
    setNoticeResetKey(resetKey);
    setPrunedNotice(null);
  }
  // Names of nodes seen in the path, so a pruned node can still be named in
  // the notice after its options have been replaced by a new date range.
  const seenNames = useRef(new Map<string, string>());

  const path: PathNode<TLevelKey>[] = [];
  for (let i = 0; i < ids.length; i++) {
    const node = levels[i]?.options.find((o) => o.id === ids[i]);
    if (!node) break;
    path.push({ ...node, level: levels[i].key });
  }

  useEffect(() => {
    for (const node of path) seenNames.current.set(`${node.level}:${node.id}`, node.name);
  });

  // Seed the selection with the first level's first option once it loads.
  // This synchronizes the URL with server data that only becomes available
  // after mount — one of the recognized effect use cases, not the "derive
  // state from a prop" anti-pattern the set-state-in-effect rule targets.
  useEffect(() => {
    const firstLevelOptions = levels[0]?.options ?? [];
    if (ids.length === 0 && firstLevelOptions.length > 0) {
      setIds([firstLevelOptions[0].id]);
    }
  }, [ids.length, levels, setIds]);

  // If the date range changes (or a shared link carries a stale id) and a
  // selected node no longer exists at its level, drop it (and everything
  // below it) and surface why. Same justification as above.
  useEffect(() => {
    if (ids.length === 0) return;

    for (let i = 0; i < ids.length; i++) {
      const level = levels[i];
      if (!level.isSuccess) return;
      if (!level.options.some((o) => o.id === ids[i])) {
        const droppedName = seenNames.current.get(`${level.key}:${ids[i]}`) ?? level.label;
        setPrunedNotice(droppedName);

        setIds(ids.slice(0, i));
        return;
      }
    }
  }, [ids, levels, setIds]);

  const handleSelect = (levelIndex: number, idStr: string) => {
    const node = (levels[levelIndex]?.options ?? []).find((o) => String(o.id) === idStr);
    if (!node) return;
    setPrunedNotice(null);
    setIds([...ids.slice(0, levelIndex), node.id]);
  };

  const handleAddLevel = () => {
    if (ids.length >= levels.length) return;
    const children = levels[ids.length]?.options ?? [];
    if (children.length === 0) return;
    setPrunedNotice(null);
    setIds([...ids, children[0].id]);
  };

  const handleRemoveLevel = () => {
    if (ids.length <= 1) return;
    setPrunedNotice(null);
    setIds(ids.slice(0, -1));
  };

  return { path, prunedNotice, handleSelect, handleAddLevel, handleRemoveLevel };
}
