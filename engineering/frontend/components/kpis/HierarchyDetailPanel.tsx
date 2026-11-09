"use client";

import { useState, type FormEvent } from "react";
import { HIERARCHIES, type HierarchyId } from "@/lib/kpi-config";
import { getErrorMessage } from "@/lib/errors";
import { PERMISSION_MESSAGES } from "@/lib/action-permissions";
import { CRUD_ACTIONS } from "@/lib/scopes";
import { THEME } from "@/lib/theme";
import type { HierarchyRow } from "@/types/kpi";
import type { HierarchyLevelMutations, HierarchySelection } from "@/types/kpi-admin";
import { nodeKey } from "@/components/kpis/HierarchyTreeNode";
import { DISABLED_INPUT_CLASS, PermissionsHint, usePermissionUi } from "@/components/kpis/permission-ui";

interface HierarchyDetailPanelProps {
  hierarchyId: HierarchyId;
  selection: HierarchySelection;
  mutations: HierarchyLevelMutations;
  /** Whether to explain disabled actions here (the tree header already does it for the top level). */
  showPermissionsHint: boolean;
  onSelectAncestor: (ancestor: HierarchySelection) => void;
  onUpdated: (row: HierarchyRow) => void;
}

const INPUT_CLASS =
  "w-full bg-gray-50 rounded-lg p-2.5 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0";
const LABEL_CLASS = "text-[10px] font-black uppercase text-gray-400 tracking-wider block mb-1.5";

export default function HierarchyDetailPanel({
  hierarchyId,
  selection,
  mutations,
  showPermissionsHint,
  onSelectAncestor,
  onUpdated,
}: HierarchyDetailPanelProps) {
  const levelDef = HIERARCHIES[hierarchyId].levels[selection.levelIndex];

  return (
    <div className="space-y-5 p-1">
      <div>
        <span className={`inline-block text-[10px] font-black uppercase tracking-wider px-2 py-1 rounded-md ${THEME.chip}`}>
          {levelDef.title}
        </span>
        {selection.ancestors.length > 0 && (
          <nav aria-label="Breadcrumb" className="mt-2 flex flex-wrap items-center gap-1 text-[11px] text-gray-400">
            {selection.ancestors.map((ancestor) => (
              <span key={nodeKey(ancestor.levelIndex, ancestor.row.id)} className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => onSelectAncestor(ancestor)}
                  className="cursor-pointer font-semibold text-gray-500 hover:text-gray-900 hover:underline"
                >
                  {ancestor.row.code} — {ancestor.row.name}
                </button>
                <span aria-hidden="true">›</span>
              </span>
            ))}
          </nav>
        )}
      </div>

      <EditForm
        key={nodeKey(selection.levelIndex, selection.row.id)}
        itemLabel={levelDef.itemLabel}
        row={selection.row}
        mutations={mutations}
        showPermissionsHint={showPermissionsHint}
        onUpdated={onUpdated}
      />
    </div>
  );
}

function EditForm({
  itemLabel,
  row,
  mutations,
  showPermissionsHint,
  onUpdated,
}: {
  itemLabel: string;
  row: HierarchyRow;
  mutations: HierarchyLevelMutations;
  showPermissionsHint: boolean;
  onUpdated: (row: HierarchyRow) => void;
}) {
  const [name, setName] = useState(row.name);
  const [code, setCode] = useState(row.code);
  const [error, setError] = useState<string | null>(null);
  const { updateMut, parentField, permissions } = mutations;
  const { actionProps, inputProps, hintProps } = usePermissionUi(permissions, itemLabel);
  const isDirty = name.trim() !== row.name || code.trim() !== row.code;

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    // The inputs and Save are already disabled without permission; this is a backstop.
    if (!permissions.allowed[CRUD_ACTIONS.update]) {
      return setError(PERMISSION_MESSAGES.denied([CRUD_ACTIONS.update], itemLabel));
    }
    const trimmedName = name.trim();
    const trimmedCode = code.trim();

    if (!trimmedName) return setError(`${itemLabel} name is required.`);
    if (!trimmedCode) return setError(`${itemLabel} code is required.`);
    setError(null);

    const payload: Record<string, unknown> = { name: trimmedName, code: trimmedCode };
    if (parentField && row.parentId != null) payload[parentField] = row.parentId;

    updateMut.mutate(
      { id: row.id, payload },
      {
        onSuccess: () => onUpdated({ ...row, name: trimmedName, code: trimmedCode }),
        onError: (err) => setError(getErrorMessage(err, "Failed to update item.")),
      },
    );
  }

  const save = actionProps(CRUD_ACTIONS.update, updateMut.isPending);

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      {showPermissionsHint && <PermissionsHint {...hintProps} />}
      <div className="space-y-3">
        <div>
          <label htmlFor="edit-item-name" className={LABEL_CLASS}>
            {itemLabel} Name
          </label>
          <input
            id="edit-item-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            {...inputProps(CRUD_ACTIONS.update)}
            className={`${INPUT_CLASS} ${DISABLED_INPUT_CLASS}`}
          />
        </div>
        <div>
          <label htmlFor="edit-item-code" className={LABEL_CLASS}>
            Code
          </label>
          <input
            id="edit-item-code"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            {...inputProps(CRUD_ACTIONS.update)}
            className={`${INPUT_CLASS} ${DISABLED_INPUT_CLASS}`}
          />
        </div>
      </div>
      {error && <p className="text-[11px] font-bold text-red-500">{error}</p>}
      <div className="flex justify-end gap-2">
        {isDirty && (
          <button
            type="button"
            onClick={() => {
              setName(row.name);
              setCode(row.code);
              setError(null);
            }}
            className="cursor-pointer text-xs font-bold text-gray-500 bg-gray-100 hover:bg-gray-200 px-4 py-2 rounded-full transition-all"
          >
            Reset
          </button>
        )}
        <button
          type="submit"
          {...save}
          disabled={save.disabled || !isDirty}
          className={`cursor-pointer ${THEME.button} text-xs font-bold px-5 py-2 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
        >
          {updateMut.isPending ? "Saving..." : "Save"}
        </button>
      </div>
    </form>
  );
}
