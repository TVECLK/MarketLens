"use client";

import SourceDialog from "@/components/sources/SourceDialog";
import { THEME } from "@/lib/theme";
import type { Source } from "@/types/source";

interface DisableSourceModalProps {
  source: Source;
  isPending: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

/** Confirms turning a source off, explaining what that does for its type. */
export default function DisableSourceModal({ source, isPending, onCancel, onConfirm }: DisableSourceModalProps) {
  return (
    <SourceDialog
      title={`Disable "${source.name}"?`}
      onClose={onCancel}
      footer={
        <>
          <button
            type="button"
            onClick={onCancel}
            autoFocus
            className={`cursor-pointer text-xs font-bold ${THEME.buttonSecondary} px-4 py-2 rounded-full transition-all`}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isPending}
            className="cursor-pointer text-xs font-bold text-white bg-amber-600 hover:bg-amber-700 px-4 py-2 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isPending ? "Disabling..." : "Disable"}
          </button>
        </>
      }
    >
      {source.type === "auto" ? (
        <ul className="list-disc pl-4 space-y-1">
          <li>
            Automatic collection from <strong className="text-gray-800">{source.name}</strong> stops from the next
            run, and any jobs still collected for it are rejected.
          </li>
          <li>Its open jobs stay open until you enable it again.</li>
        </ul>
      ) : (
        <ul className="list-disc pl-4 space-y-1">
          <li>
            New uploads from <strong className="text-gray-800">{source.name}</strong> will be rejected.
          </li>
          <li>Existing jobs are kept.</li>
        </ul>
      )}
      <p>You can enable it again at any time.</p>
    </SourceDialog>
  );
}
