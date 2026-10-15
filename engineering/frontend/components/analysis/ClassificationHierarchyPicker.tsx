"use client";

import type {
  HierarchyLevelState,
  PathNode,
} from "@/hooks/use-hierarchy-drilldown";

interface ClassificationHierarchyPickerProps<TLevelKey extends string> {
  levels: HierarchyLevelState<TLevelKey>[];
  path: PathNode<TLevelKey>[];
  isInitializing: boolean;
  initializingLabel: string;
  initialLoadError: boolean;
  initialLoadErrorLabel: string;
  initialLoadErrorMessage?: string;
  emptyLabel: string;
  prunedNotice: string | null;
  onSelect: (levelIndex: number, idStr: string) => void;
  onAddLevel: () => void;
  onRemoveLevel: () => void;
}

export default function ClassificationHierarchyPicker<TLevelKey extends string>({
  levels,
  path,
  isInitializing,
  initializingLabel,
  initialLoadError,
  initialLoadErrorLabel,
  initialLoadErrorMessage,
  emptyLabel,
  prunedNotice,
  onSelect,
  onAddLevel,
  onRemoveLevel,
}: ClassificationHierarchyPickerProps<TLevelKey>) {
  const nextLevel = levels[path.length];

  return (
    <div className="bg-white border border-zinc-200 p-5 rounded-xl shadow-sm">
      <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 mb-4">
        Classification Hierarchy
      </h4>

      {isInitializing ? (
        <p className="text-xs text-zinc-400">{initializingLabel}</p>
      ) : initialLoadError ? (
        <p className="text-xs text-red-500 font-bold">
          {initialLoadErrorLabel}
          {initialLoadErrorMessage ? `: ${initialLoadErrorMessage}` : "."}
        </p>
      ) : (levels[0]?.options.length ?? 0) === 0 ? (
        <p className="text-xs text-zinc-400">{emptyLabel}</p>
      ) : (
        <div className="flex flex-wrap items-end gap-2">
          {path.map((node, levelIndex) => {
            const level = levels[levelIndex];
            return (
              <div key={levelIndex} className="flex items-end gap-2">
                {levelIndex > 0 && (
                  <span className="pb-2.5 text-zinc-300 font-black select-none" aria-hidden>
                    →
                  </span>
                )}
                <div>
                  <p className="text-[10px] font-black uppercase tracking-wider text-zinc-400 mb-1">
                    {level.label}
                  </p>
                  <select
                    value={node.id}
                    title={node.name}
                    disabled={level.isLoading}
                    onChange={(e) => onSelect(levelIndex, e.target.value)}
                    className="cursor-pointer bg-zinc-100 rounded-xl px-3 py-2.5 text-xs font-bold text-zinc-700 outline-none border border-transparent focus:border-zinc-300 max-w-[200px] truncate disabled:opacity-50"
                  >
                    {level.options.map((opt) => (
                      <option key={opt.id} value={opt.id}>
                        {opt.name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            );
          })}

          {path.length < levels.length && (
            <button
              onClick={onAddLevel}
              disabled={!nextLevel || nextLevel.isLoading || nextLevel.options.length === 0}
              aria-label={`Add ${nextLevel?.label}`}
              title={`Add ${nextLevel?.label}`}
              className="cursor-pointer mb-0.5 w-9 h-9 rounded-xl bg-zinc-900 hover:bg-zinc-700 text-white font-black text-base flex items-center justify-center transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              +
            </button>
          )}

          {path.length > 1 && (
            <button
              onClick={onRemoveLevel}
              aria-label="Remove last level"
              title="Remove last level"
              className="cursor-pointer mb-0.5 w-9 h-9 rounded-xl bg-zinc-100 hover:bg-zinc-200 text-zinc-500 font-black text-base flex items-center justify-center transition-colors"
            >
              −
            </button>
          )}
        </div>
      )}
      {prunedNotice && (
        <p className="text-[11px] text-amber-600 font-bold mt-3">
          {prunedNotice} has no data in this date range, showing the closest level
          with data.
        </p>
      )}
    </div>
  );
}
