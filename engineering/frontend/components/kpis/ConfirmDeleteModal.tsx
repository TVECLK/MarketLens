import type { ReactNode } from "react";

interface ConfirmDeleteModalProps {
  title: string;
  children: ReactNode;
  onCancel: () => void;
  onConfirm: () => void;
  isPending: boolean;
}

export default function ConfirmDeleteModal({
  title,
  children,
  onCancel,
  onConfirm,
  isPending,
}: ConfirmDeleteModalProps) {
  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 px-4">
      <div className="bg-white rounded-xl shadow-xl max-w-sm w-full p-6">
        <h3 className="text-sm font-black text-gray-900">{title}</h3>
        <div className="text-xs text-gray-500 mt-2">{children}</div>
        <div className="flex justify-end gap-2 mt-6">
          <button
            onClick={onCancel}
            className="text-xs font-bold text-gray-500 bg-gray-100 hover:bg-gray-200 px-4 py-2 rounded-lg transition-all"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={isPending}
            className="text-xs font-bold text-white bg-red-600 hover:bg-red-700 px-4 py-2 rounded-lg transition-all disabled:opacity-50"
          >
            {isPending ? "Deleting..." : "Delete"}
          </button>
        </div>
      </div>
    </div>
  );
}
