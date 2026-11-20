import type { ReactNode } from "react";
import { clsx } from "clsx";

interface LoadingSpinnerProps {
  /** Text shown with the spinner, e.g. "Loading skills...". */
  label: ReactNode;
  /** `md` (24px) for content areas, `sm` (14px) for a single line inside a list or form. */
  size?: "sm" | "md";
  /** Spinner beside the text on one line, instead of above it. */
  inline?: boolean;
  className?: string;
}

/**
 * The loading indicator for areas inside a page: a spinning ring with its text. Same ring style as
 * the full-page loading screens. Text colour and size come from `className` or the parent.
 */
export default function LoadingSpinner({ label, size = "md", inline = false, className }: LoadingSpinnerProps) {
  return (
    <div
      role="status"
      aria-live="polite"
      className={clsx(
        inline ? "inline-flex items-center gap-2" : "flex flex-col items-center justify-center gap-3",
        className,
      )}
    >
      <span
        aria-hidden
        className={clsx(
          "shrink-0 rounded-full border-2 border-zinc-200 border-t-zinc-500 animate-spin",
          size === "sm" ? "w-3.5 h-3.5" : "w-6 h-6",
        )}
      />
      {label}
    </div>
  );
}
