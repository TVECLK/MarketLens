"use client";

import { useEffect, useId, type ReactNode } from "react";
import { XMarkIcon } from "@heroicons/react/16/solid";

interface SourceDialogProps {
  title: ReactNode;
  children: ReactNode;
  /** The buttons, shown at the bottom right. */
  footer: ReactNode;
  onClose: () => void;
  /** Wraps the content in a form, so Enter submits it. */
  onSubmit?: () => void;
  size?: "sm" | "md";
}

/** The overlay and card every source dialog uses, styled like the KPI modals. Escape closes it. */
export default function SourceDialog({ title, children, footer, onClose, onSubmit, size = "sm" }: SourceDialogProps) {
  const titleId = useId();

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  const body = (
    <>
      <div className="flex items-start justify-between gap-4">
        <h3 id={titleId} className="text-sm font-black text-gray-900">
          {title}
        </h3>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="cursor-pointer -mt-1 -mr-1 p-1 rounded-full text-gray-400 hover:text-gray-700 hover:bg-gray-100 transition-all"
        >
          <XMarkIcon className="w-4 h-4" />
        </button>
      </div>
      <div className="text-xs text-gray-500 mt-3 space-y-3">{children}</div>
      <div className="flex justify-end gap-2 mt-6">{footer}</div>
    </>
  );

  const cardClass = `bg-white rounded-xl shadow-xl w-full p-6 ${size === "md" ? "max-w-lg" : "max-w-sm"}`;

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 px-4">
      {onSubmit ? (
        <form
          role="dialog"
          aria-modal="true"
          aria-labelledby={titleId}
          className={cardClass}
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit();
          }}
        >
          {body}
        </form>
      ) : (
        <div role="dialog" aria-modal="true" aria-labelledby={titleId} className={cardClass}>
          {body}
        </div>
      )}
    </div>
  );
}
