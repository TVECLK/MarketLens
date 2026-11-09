"use client";

import { useState, type FormEvent } from "react";
import { getErrorMessage } from "@/lib/errors";
import { PERMISSION_MESSAGES } from "@/lib/action-permissions";
import { CRUD_ACTIONS } from "@/lib/scopes";
import { THEME } from "@/lib/theme";
import type { HierarchyRow } from "@/types/kpi";
import type { HierarchyLevelMutations } from "@/types/kpi-admin";

interface CreateItemModalProps {
  itemLabel: string;
  /** The item the new one is created under; null when creating a top-level item. */
  parent: HierarchyRow | null;
  parentLabel: string | null;
  mutations: HierarchyLevelMutations;
  onCancel: () => void;
  onCreated: (name: string) => void;
}


export default function CreateItemModal({
  itemLabel,
  parent,
  parentLabel,
  mutations,
  onCancel,
  onCreated,
}: CreateItemModalProps) {
  const [name, setName] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState<string | null>(null);
  const { createMut, parentField, permissions } = mutations;

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    // The Add buttons are already disabled without permission; this is a backstop.
    if (!permissions.allowed[CRUD_ACTIONS.create]) {
      return setError(PERMISSION_MESSAGES.denied([CRUD_ACTIONS.create], itemLabel));
    }

    const trimmedName = name.trim();
    const trimmedCode = code.trim();

    if (!trimmedName) return setError(`${itemLabel} name is required.`);
    if (!trimmedCode) return setError(`${itemLabel} code is required.`);
    setError(null);

    const payload: Record<string, unknown> = { name: trimmedName, code: trimmedCode };
    if (parentField && parent) payload[parentField] = parent.id;

    createMut.mutate(payload, {
      onSuccess: () => onCreated(trimmedName),
      onError: (err) => setError(getErrorMessage(err, "Failed to create item.")),
    });
  }

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 px-4">
      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-xl max-w-md w-full p-6 space-y-4">
        <h3 className="text-sm font-black text-gray-900">Add {itemLabel}</h3>

        {parent && parentLabel && (
          <div>
            <span className={THEME.form.label}>Parent {parentLabel}</span>
            <p className="text-xs font-bold text-gray-700 bg-gray-100 rounded-lg p-2.5">
              {parent.code} — {parent.name}
            </p>
          </div>
        )}

        <div>
          <label htmlFor="create-item-name" className={THEME.form.label}>
            {itemLabel} Name
          </label>
          <input
            id="create-item-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={`e.g. New ${itemLabel}`}
            className={THEME.form.input}
            autoFocus
          />
        </div>
        <div>
          <label htmlFor="create-item-code" className={THEME.form.label}>
            Code
          </label>
          <input
            id="create-item-code"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="e.g. 1112"
            className={THEME.form.input}
          />
        </div>

        {error && <p className="text-[11px] font-bold text-red-500">{error}</p>}

        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onCancel}
            className={`cursor-pointer text-xs font-bold ${THEME.buttonSecondary} px-4 py-2 rounded-full transition-all`}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createMut.isPending}
            className={`cursor-pointer text-xs font-bold ${THEME.button} px-4 py-2 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {createMut.isPending ? "Adding..." : `Add ${itemLabel}`}
          </button>
        </div>
      </form>
    </div>
  );
}
