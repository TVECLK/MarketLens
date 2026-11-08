import type { FormEvent } from "react";
import Link from "next/link";
import type { HierarchyDef, AccentClasses } from "@/lib/kpi-config";
import type { HierarchyLevelBundle } from "@/types/kpi-admin";
import type { HierarchyRow } from "@/types/kpi";
import { CRUD_ACTIONS } from "@/lib/scopes";
import {
  DISABLED_INPUT_CLASS,
  DISABLED_ROW_ACTION_CLASS,
  PermissionsHint,
  usePermissionUi,
} from "@/components/kpis/permission-ui";

interface HierarchyLevelPanelProps {
  hierarchyDef: HierarchyDef;
  hAccent: AccentClasses;
  activeLevelIndex: number;
  levelCounts: number[];
  levelHref: (index: number) => string;
  pageSize: number;

  levelBundle: HierarchyLevelBundle;
  parentItems: HierarchyRow[];

  newName: string;
  onNewNameChange: (value: string) => void;
  newCode: string;
  onNewCodeChange: (value: string) => void;
  newParentId: string;
  onNewParentIdChange: (value: string) => void;
  formError: string | null;
  onSubmit: (e: FormEvent) => void;

  editingId: number | null;
  editName: string;
  onEditNameChange: (value: string) => void;
  editCode: string;
  onEditCodeChange: (value: string) => void;
  editParentId: string;
  onEditParentIdChange: (value: string) => void;
  editError: string | null;
  onStartEdit: (item: HierarchyRow) => void;
  onCancelEdit: () => void;
  onSaveEdit: (id: number) => void;
  onRequestDelete: (item: HierarchyRow) => void;
}

export default function HierarchyLevelPanel({
  hierarchyDef,
  hAccent,
  activeLevelIndex,
  levelCounts,
  levelHref,
  pageSize,
  levelBundle,
  parentItems,
  newName,
  onNewNameChange,
  newCode,
  onNewCodeChange,
  newParentId,
  onNewParentIdChange,
  formError,
  onSubmit,
  editingId,
  editName,
  onEditNameChange,
  editCode,
  onEditCodeChange,
  editParentId,
  onEditParentIdChange,
  editError,
  onStartEdit,
  onCancelEdit,
  onSaveEdit,
  onRequestDelete,
}: HierarchyLevelPanelProps) {
  const levelDef = hierarchyDef.levels[activeLevelIndex];
  const parentLevelDef = activeLevelIndex > 0 ? hierarchyDef.levels[activeLevelIndex - 1] : null;
  const items = levelBundle.items;
  const { allowed, actionProps, inputProps, hintProps, emptyMessage } = usePermissionUi(levelBundle.permissions, levelDef.itemLabel);

  return (
    <main className="flex-1 p-4 mx-auto space-y-4">
      <div>
        <h2 className="text-lg font-black text-gray-900">{hierarchyDef.label} Classification</h2>
        <p className="text-xs text-gray-400 mt-1">
          {hierarchyDef.levels.map((l) => l.title).join(" → ")}. Each item has a name, a code, and
          (except the top level) a parent from the level above.
        </p>
      </div>

      <div className="flex flex-wrap gap-2">
        {hierarchyDef.levels.map((level, idx) => {
          const isActive = activeLevelIndex === idx;
          return (
            <Link
              key={level.key}
              href={levelHref(idx)}
              aria-current={isActive ? "page" : undefined}
              className={`flex items-center gap-2 px-4 py-2 rounded-md text-xs font-bold transition-all ${isActive ? `${hAccent.solidBg} ${hAccent.solidText}` : "bg-white text-gray-500 hover:bg-gray-100 shadow-sm"
                }`}
            >
              <span
                className={`w-4 h-4 rounded-full flex items-center justify-center text-[9px] ${isActive ? "bg-white/70" : "bg-gray-100"
                  }`}
              >
                {idx + 1}
              </span>
              {level.title}
              <span className={isActive ? "" : "text-gray-300"}>({levelCounts[idx]})</span>
            </Link>
          );
        })}
      </div>

      {levelBundle.isError && (
        <p className="text-xs font-bold text-red-500">
          Failed to load {levelDef.title.toLowerCase()} data. Please refresh.
        </p>
      )}

      <PermissionsHint {...hintProps} />

      <form
        onSubmit={onSubmit}
        className="bg-white p-3 rounded-md shadow-sm flex flex-col sm:flex-row gap-3 sm:items-end sm:flex-wrap"
      >
        <div className="flex-1 min-w-[160px]">
          <label className="text-[10px] font-black uppercase text-gray-400 tracking-wider block mb-1.5">
            {levelDef.itemLabel} Name
          </label>
          <input
            value={newName}
            onChange={(e) => onNewNameChange(e.target.value)}
            placeholder={`e.g. New ${levelDef.itemLabel}`}
            {...inputProps(CRUD_ACTIONS.create)}
            className={`w-full bg-gray-50 rounded-lg p-2.5 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0 ${DISABLED_INPUT_CLASS}`}
          />
        </div>
        <div className="w-full sm:w-32">
          <label className="text-[10px] font-black uppercase text-gray-400 tracking-wider block mb-1.5">
            Code
          </label>
          <input
            value={newCode}
            onChange={(e) => onNewCodeChange(e.target.value)}
            placeholder="e.g. 1112"
            {...inputProps(CRUD_ACTIONS.create)}
            className={`w-full bg-gray-50 rounded-lg p-2.5 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0 ${DISABLED_INPUT_CLASS}`}
          />
        </div>
        {parentLevelDef && (
          <div className="w-full sm:w-64">
            <label className="text-[10px] font-black uppercase text-gray-400 tracking-wider block mb-1.5">
              Parent {parentLevelDef.itemLabel}
            </label>
            <select
              value={newParentId}
              onChange={(e) => onNewParentIdChange(e.target.value)}
              {...inputProps(CRUD_ACTIONS.create)}
              className={`w-full bg-gray-50 rounded-lg p-2.5 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0 ${DISABLED_INPUT_CLASS}`}
            >
              <option value="">Select {parentLevelDef.itemLabel}…</option>
              {parentItems.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.code} — {p.name}
                </option>
              ))}
            </select>
          </div>
        )}
        <button
          type="submit"
          {...actionProps(CRUD_ACTIONS.create, levelBundle.createMut.isPending)}
          className={`${hAccent.bg} ${hAccent.solidText} text-xs font-bold px-5 py-2.5 rounded-lg hover:opacity-90 transition-all shrink-0 disabled:opacity-50 disabled:cursor-not-allowed`}
        >
          {levelBundle.createMut.isPending ? "Adding..." : `+ Add ${levelDef.itemLabel}`}
        </button>
      </form>
      {allowed[CRUD_ACTIONS.create] && parentLevelDef && parentItems.length === 0 && (
        <p className="text-[11px] font-bold text-amber-600 -mt-3 px-1">
          Add at least one {parentLevelDef.itemLabel} first before creating a{" "}
          {levelDef.itemLabel.toLowerCase()}.
        </p>
      )}
      {formError && <p className="text-[11px] font-bold text-red-500 -mt-3 px-1">{formError}</p>}

      <div className="bg-white rounded-md shadow-sm overflow-hidden">
        <table className="w-full text-xs">
          <thead>
            <tr className="bg-gray-50 text-gray-400 text-[10px] font-black uppercase tracking-wider">
              <th className="text-left px-5 py-3">{levelDef.itemLabel}</th>
              <th className="text-left px-5 py-3 w-32">Code</th>
              {parentLevelDef && (
                <th className="text-left px-5 py-3 w-64">Parent {parentLevelDef.itemLabel}</th>
              )}
              <th className="text-right px-5 py-3 w-40">Actions</th>
            </tr>
          </thead>
          <tbody>
            {levelBundle.isLoading && (
              <tr>
                <td colSpan={parentLevelDef ? 4 : 3} className="px-5 py-10 text-center text-gray-400 font-medium">
                  Loading...
                </td>
              </tr>
            )}
            {!levelBundle.isLoading && items.length === 0 && (
              <tr>
                <td colSpan={parentLevelDef ? 4 : 3} className="px-5 py-10 text-center text-gray-400 font-medium">
                  {emptyMessage}
                </td>
              </tr>
            )}
            {items.map((item) => {
              const isEditing = editingId === item.id;
              return (
                <tr key={item.id} className="border-t border-gray-50 hover:bg-gray-50/60">
                  <td className="px-5 py-3 align-top">
                    {isEditing ? (
                      <input
                        value={editName}
                        onChange={(e) => onEditNameChange(e.target.value)}
                        className="w-full bg-gray-50 rounded-lg p-2 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0"
                        autoFocus
                      />
                    ) : (
                      <span className="font-bold text-gray-800">{item.name}</span>
                    )}
                  </td>
                  <td className="px-5 py-3 align-top">
                    {isEditing ? (
                      <input
                        value={editCode}
                        onChange={(e) => onEditCodeChange(e.target.value)}
                        className="w-full bg-gray-50 rounded-lg p-2 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0"
                      />
                    ) : (
                      <span className="font-mono font-bold text-gray-600">{item.code}</span>
                    )}
                  </td>
                  {parentLevelDef && (
                    <td className="px-5 py-3 align-top">
                      {isEditing ? (
                        <select
                          value={editParentId}
                          onChange={(e) => onEditParentIdChange(e.target.value)}
                          className="w-full bg-gray-50 rounded-lg p-2 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0"
                        >
                          <option value="">Select {parentLevelDef.itemLabel}…</option>
                          {parentItems.map((p) => (
                            <option key={p.id} value={p.id}>
                              {p.code} — {p.name}
                            </option>
                          ))}
                        </select>
                      ) : (
                        <span className="text-gray-500">{item.parentLabel ?? "—"}</span>
                      )}
                    </td>
                  )}
                  <td className="px-5 py-3 align-top">
                    {isEditing && editError && (
                      <p className="text-[10px] font-bold text-red-500 mb-1 text-right">{editError}</p>
                    )}
                    <div className="flex justify-end gap-2">
                      {isEditing ? (
                        <>
                          <button
                            onClick={() => onSaveEdit(item.id)}
                            {...actionProps(CRUD_ACTIONS.update, levelBundle.updateMut.isPending)}
                            className="text-[11px] font-bold text-white bg-emerald-600 hover:bg-emerald-700 px-3 py-1.5 rounded-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            {levelBundle.updateMut.isPending ? "Saving..." : "Save"}
                          </button>
                          <button
                            onClick={onCancelEdit}
                            className="text-[11px] font-bold text-gray-500 bg-gray-100 hover:bg-gray-200 px-3 py-1.5 rounded-lg transition-all"
                          >
                            Cancel
                          </button>
                        </>
                      ) : (
                        <>
                          <button
                            onClick={() => onStartEdit(item)}
                            {...actionProps(CRUD_ACTIONS.update)}
                            className={`text-[11px] font-bold ${hAccent.text} ${hAccent.lightBg} hover:opacity-80 px-3 py-1.5 rounded-lg transition-all ${DISABLED_ROW_ACTION_CLASS}`}
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => onRequestDelete(item)}
                            {...actionProps(CRUD_ACTIONS.delete)}
                            className={`text-[11px] font-bold text-red-500 bg-red-50 hover:bg-red-100 px-3 py-1.5 rounded-lg transition-all ${DISABLED_ROW_ACTION_CLASS}`}
                          >
                            Delete
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {levelBundle.total !== undefined && levelBundle.setPage && (
        <div className="flex items-center justify-between px-1">
          <p className="text-xs text-gray-400">
            Page {levelBundle.page} of {Math.max(1, Math.ceil(levelBundle.total / pageSize))} (
            {levelBundle.total} total)
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => levelBundle.setPage!(Math.max(1, (levelBundle.page ?? 1) - 1))}
              disabled={levelBundle.page === 1}
              className="px-3 py-1.5 text-xs font-bold rounded-lg border border-gray-200 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Previous
            </button>
            <button
              onClick={() => levelBundle.setPage!((levelBundle.page ?? 1) + 1)}
              disabled={(levelBundle.page ?? 1) >= Math.ceil(levelBundle.total / pageSize)}
              className="px-3 py-1.5 text-xs font-bold rounded-lg border border-gray-200 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </main>
  );
}
