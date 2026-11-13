import { PencilIcon, PlusIcon, TrashIcon } from "@heroicons/react/16/solid";
import { DISABLED_ROW_ACTION_CLASS } from "@/components/kpis/permission-ui";

export type RowActionKind = "add" | "edit" | "delete";

const KIND_CLASSES: Record<RowActionKind, string> = {
  add: "text-gray-400 hover:text-gray-900 hover:bg-gray-100",
  edit: "text-gray-400 hover:text-gray-900 hover:bg-gray-100",
  delete: "text-red-500 hover:text-red-700 hover:bg-red-50",
};

const ICONS: Record<RowActionKind, typeof PlusIcon> = {
  add: PlusIcon,
  edit: PencilIcon,
  delete: TrashIcon,
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
  const Icon = ICONS[kind];
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
      <Icon className="w-3.5 h-3.5" />
    </button>
  );
}
