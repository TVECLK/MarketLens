"use client";

import { useState } from "react";
import SourceDialog from "@/components/sources/SourceDialog";
import { confirmsSourceName, displaySourceName } from "@/lib/source-name";
import { THEME } from "@/lib/theme";
import type { Source } from "@/types/source";

interface DeleteSourceModalProps {
  source: Source;
  isPending: boolean;
  error: string | null;
  onCancel: () => void;
  onConfirm: () => void;
}

/**
 * Confirms deleting a source. A manual source needs one click. A crawler (auto) source affects a
 * running crawler, so the admin must type its exact name, and is pointed to disabling instead.
 */
export default function DeleteSourceModal({ source, isPending, error, onCancel, onConfirm }: DeleteSourceModalProps) {
  const [typed, setTyped] = useState("");
  const isAuto = source.type === "auto";
  const displayName = displaySourceName(source.name);
  const canConfirm = !isPending && (!isAuto || confirmsSourceName(typed, source.name));

  return (
    <SourceDialog
      title={isAuto ? `Delete automatic source "${displayName}"?` : `Delete "${displayName}"?`}
      onClose={onCancel}
      onSubmit={() => canConfirm && onConfirm()}
      size={isAuto ? "md" : "sm"}
      footer={
        <>
          <button
            type="button"
            onClick={onCancel}
            className={`cursor-pointer text-xs font-bold ${THEME.buttonSecondary} px-4 py-2 rounded-full transition-all`}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={!canConfirm}
            className="cursor-pointer text-xs font-bold text-white bg-red-600 hover:bg-red-700 px-4 py-2 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isPending ? "Deleting..." : "Delete"}
          </button>
        </>
      }
    >
      {isAuto ? (
        <>
          <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-red-800 space-y-1.5">
            <p className="font-bold">This affects automatic collection.</p>
            <ul className="list-disc pl-4 space-y-1">
              <li>
                Jobs collected for {displayName} will be <strong>rejected</strong> until a source named{" "}
                <code className="font-mono bg-white/70 px-1 rounded">{displayName}</code> exists again.
              </li>
              <li>
                Its open jobs <strong>stay open</strong> and won&apos;t be closed automatically.
              </li>
            </ul>
          </div>
          <p>
            To stop it for a while, <strong className="text-gray-800">disable</strong> it instead: that keeps the
            source and can be undone with one click.
          </p>
          <div>
            <label htmlFor="confirm-source-name" className={THEME.form.label}>
              Type <span className="font-mono normal-case text-gray-700">{displayName}</span> to confirm
            </label>
            <input
              id="confirm-source-name"
              value={typed}
              onChange={(e) => setTyped(e.target.value)}
              autoComplete="off"
              spellCheck={false}
              className={`${THEME.form.input} font-mono`}
            />
          </div>
        </>
      ) : (
        <ul className="list-disc pl-4 space-y-1">
          <li>New uploads from this source will be rejected.</li>
          <li>Existing jobs are kept and stay linked to it.</li>
          <li>Adding a source with the same name later restores it.</li>
        </ul>
      )}
      {error && <p className="text-[11px] font-bold text-red-500">{error}</p>}
    </SourceDialog>
  );
}
