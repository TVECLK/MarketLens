interface SourceStatusSwitchProps {
  checked: boolean;
  /** The source's name, read out by screen readers. */
  label: string;
  onChange: (next: boolean) => void;
  disabled?: boolean;
  /** Replaces the tooltip, e.g. with the reason the switch is disabled. */
  title?: string;
  "aria-describedby"?: string;
}

/** An on/off switch for a source's active state. */
export default function SourceStatusSwitch({
  checked,
  label,
  onChange,
  disabled,
  title,
  "aria-describedby": describedBy,
}: SourceStatusSwitchProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={`${label} active`}
      aria-describedby={describedBy}
      title={title ?? (checked ? `Disable ${label}` : `Enable ${label}`)}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={`cursor-pointer relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-300 disabled:cursor-not-allowed disabled:opacity-50 ${
        checked ? "bg-emerald-500" : "bg-gray-300"
      }`}
    >
      <span
        aria-hidden="true"
        className={`inline-block h-4 w-4 rounded-full bg-white shadow transition-transform ${
          checked ? "translate-x-4.5" : "translate-x-0.5"
        }`}
      />
    </button>
  );
}
