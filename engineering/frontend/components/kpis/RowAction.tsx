import type { ReactNode } from "react";
import { DISABLED_ROW_ACTION_CLASS } from "@/components/kpis/permission-ui";

export type RowActionKind = "add" | "edit" | "delete";

const KIND_CLASSES: Record<RowActionKind, string> = {
  add: "text-gray-400 hover:text-gray-900 hover:bg-gray-100",
  edit: "text-gray-400 hover:text-gray-900 hover:bg-gray-100",
  delete: "text-red-500 hover:text-red-700 hover:bg-red-50",
};

const ICON_PATHS: Record<RowActionKind, ReactNode> = {
  add: (
    <>
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </>
  ),
  edit: (
    <>
      <path d="M12 20h9" />
      <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
    </>
  ),
  delete: (
    <>
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
      <path d="M10 11v6M14 11v6" />
      <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" />
    </>
  ),
};

interface RowActionProps {
  kind: RowActionKind;
  /** The hover tooltip, e.g. "Edit". */
  label: string;
  /** The item acted on. Read out by screen readers only, never shown. */
  subject?: string;
  onClick: () => void;
  disabled?: boolean;
  /** Replaces the tooltip, e.g. with the reason the action is disabled. */
  title?: string;
  "aria-describedby"?: string;
}

/** A small icon-only button for a row's add / edit / delete action. */
export default function RowAction({
  kind,
  label,
  subject,
  onClick,
  disabled,
  title,
  "aria-describedby": describedBy,
}: RowActionProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={subject ? `${label} ${subject}` : label}
      aria-describedby={describedBy}
      title={title ?? label}
      className={`cursor-pointer w-6 h-6 shrink-0 flex items-center justify-center rounded-full transition-colors ${DISABLED_ROW_ACTION_CLASS} disabled:hover:bg-transparent ${KIND_CLASSES[kind]}`}
    >
      <svg
        className="w-3.5 h-3.5"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        {ICON_PATHS[kind]}
      </svg>
    </button>
  );
}
