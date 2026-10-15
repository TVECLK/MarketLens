"use client";

import { useEffect, useState, type Dispatch, type SetStateAction } from "react";
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

type SetPath<TLevelKey extends string> = Dispatch<SetStateAction<PathNode<TLevelKey>[]>>;

/**
 * Owns just the drilldown path's state. Call this first, before any
 * per-level react-query hooks — each level's query needs `path[i - 1]?.id`
 * as an argument, so the state has to exist before those queries run.
 */
export function useHierarchyPath<TLevelKey extends string>(): [
  PathNode<TLevelKey>[],
  SetPath<TLevelKey>,
] {
  return useState<PathNode<TLevelKey>[]>([]);
}

interface UseHierarchyDrilldownResult {
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
 * Call `useHierarchyPath` first to get `path`/`setPath`, call each level's
 * own react-query hook (Rules of Hooks requires a fixed, unconditional hook
 * per level — this can't be a loop), assemble the results into `levels`,
 * then call this hook with all three.
 */
export function useHierarchyDrilldown<TLevelKey extends string>(
  path: PathNode<TLevelKey>[],
  setPath: SetPath<TLevelKey>,
  levels: HierarchyLevelState<TLevelKey>[],
): UseHierarchyDrilldownResult {
  const [prunedNotice, setPrunedNotice] = useState<string | null>(null);

  // Seed the path with the first level's first option once it loads. This
  // synchronizes local selection state with server data that only becomes
  // available after mount — one of the recognized effect use cases, not the
  // "derive state from a prop" anti-pattern the set-state-in-effect rule
  // otherwise targets.
  useEffect(() => {
    const firstLevelOptions = levels[0]?.options ?? [];
    if (path.length === 0 && firstLevelOptions.length > 0) {
      const first = firstLevelOptions[0];
      setPath([{ ...first, level: levels[0].key }]);
    }
  }, [path.length, levels, setPath]);

  // If the date range changes and a previously-selected node no longer
  // exists at its level, drop it (and everything below it) and surface why.
  // Same justification as above — same effect use case, same hook.
  useEffect(() => {
    if (path.length === 0) return;

    for (let i = 0; i < path.length; i++) {
      const level = levels[i];
      if (!level.isSuccess) return;
      if (!level.options.some((o) => o.id === path[i].id)) {
        // eslint-disable-next-line react-hooks/set-state-in-effect -- see note above
        setPrunedNotice(path[i].name);
         
        setPath(path.slice(0, i));
        return;
      }
    }

    setPrunedNotice(null);
  }, [path, levels, setPath]);

  const handleSelect = (levelIndex: number, idStr: string) => {
    const options = levels[levelIndex]?.options ?? [];
    const node = options.find((o) => String(o.id) === idStr);
    if (!node) return;
    setPath([...path.slice(0, levelIndex), { ...node, level: levels[levelIndex].key }]);
  };

  const handleAddLevel = () => {
    if (path.length >= levels.length) return;
    const children = levels[path.length]?.options ?? [];
    if (children.length === 0) return;
    setPath([...path, { ...children[0], level: levels[path.length].key }]);
  };

  const handleRemoveLevel = () => {
    if (path.length <= 1) return;
    setPath(path.slice(0, -1));
  };

  return { prunedNotice, handleSelect, handleAddLevel, handleRemoveLevel };
}
