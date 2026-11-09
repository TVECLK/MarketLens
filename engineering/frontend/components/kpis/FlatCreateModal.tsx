"use client";

import { useState, type FormEvent } from "react";
import { getErrorMessage } from "@/lib/errors";
import { PERMISSION_MESSAGES } from "@/lib/action-permissions";
import { CRUD_ACTIONS } from "@/lib/scopes";
import { THEME } from "@/lib/theme";
import type { FlatBundle } from "@/types/kpi-admin";

interface FlatCreateModalProps {
  itemLabel: string;
  bundle: FlatBundle;
  onCancel: () => void;
  onCreated: (name: string) => void;
}

export default function FlatCreateModal({ itemLabel, bundle, onCancel, onCreated }: FlatCreateModalProps) {
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    // The Add button is already disabled without permission; this is a backstop.
    if (!bundle.permissions.allowed[CRUD_ACTIONS.create]) {
      return setError(PERMISSION_MESSAGES.denied([CRUD_ACTIONS.create], itemLabel));
    }
    const trimmed = name.trim();
    if (!trimmed) return setError(`${itemLabel} name is required.`);
    setError(null);

    bundle.createMut.mutate(
      { [bundle.labelField]: trimmed },
      {
        onSuccess: () => onCreated(trimmed),
        onError: (err) => setError(getErrorMessage(err, "Failed to create item.")),
      },
    );
  }

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 px-4">
      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-xl max-w-md w-full p-6 space-y-4">
        <h3 className="text-sm font-black text-gray-900">Add {itemLabel}</h3>

        <div>
          <label
            htmlFor="create-flat-item-name"
            className={THEME.form.label}
          >
            {itemLabel}
          </label>
          <input
            id="create-flat-item-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={`e.g. New ${itemLabel}`}
            className={THEME.form.input}
            autoFocus
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
            disabled={bundle.createMut.isPending}
            className={`cursor-pointer text-xs font-bold ${THEME.button} px-4 py-2 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {bundle.createMut.isPending ? "Adding..." : `Add ${itemLabel}`}
          </button>
        </div>
      </form>
    </div>
  );
}
