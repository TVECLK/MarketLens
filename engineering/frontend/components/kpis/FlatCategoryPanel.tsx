import type { FormEvent } from "react";
import type { KPICategoryConfig, AccentClasses } from "@/lib/kpi-config";
import type { FlatBundle, FlatItem } from "@/types/kpi-admin";

interface FlatCategoryPanelProps {
  config: KPICategoryConfig;
  accent: AccentClasses;
  bundle: FlatBundle;
  newName: string;
  onNewNameChange: (value: string) => void;
  formError: string | null;
  onSubmit: (e: FormEvent) => void;
  editingId: number | null;
  editName: string;
  onEditNameChange: (value: string) => void;
  editError: string | null;
  onStartEdit: (item: FlatItem) => void;
  onCancelEdit: () => void;
  onSaveEdit: (id: number) => void;
  onRequestDelete: (item: FlatItem) => void;
}

export default function FlatCategoryPanel({
  config,
  accent,
  bundle,
  newName,
  onNewNameChange,
  formError,
  onSubmit,
  editingId,
  editName,
  onEditNameChange,
  editError,
  onStartEdit,
  onCancelEdit,
  onSaveEdit,
  onRequestDelete,
}: FlatCategoryPanelProps) {
  const items = bundle.items;

  return (
    <main className="flex-1 p-8 max-w-4xl mx-auto space-y-6">
      <div>
        <h2 className="text-lg font-black text-gray-900">{config.title}</h2>
        <p className="text-xs text-gray-400 mt-1">{config.description}</p>
      </div>

      <div className="bg-white p-5 rounded-xl shadow-sm w-fit min-w-[200px]">
        <p className="text-[11px] text-gray-400 font-bold uppercase tracking-wider">Total Items</p>
        <p className="text-2xl font-black text-gray-900 mt-1">{items.length}</p>
      </div>

      {bundle.isError && (
        <p className="text-xs font-bold text-red-500">
          Failed to load {config.title.toLowerCase()}. Please refresh.
        </p>
      )}

      <form
        onSubmit={onSubmit}
        className="bg-white p-5 rounded-xl shadow-sm flex flex-col sm:flex-row gap-3 sm:items-end"
      >
        <div className="flex-1">
          <label className="text-[10px] font-black uppercase text-gray-400 tracking-wider block mb-1.5">
            {config.itemLabel}
          </label>
          <input
            value={newName}
            onChange={(e) => onNewNameChange(e.target.value)}
            placeholder={`e.g. New ${config.itemLabel}`}
            className="w-full bg-gray-50 rounded-lg p-2.5 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0"
          />
        </div>
        <button
          type="submit"
          disabled={bundle.createMut.isPending}
          className={`${accent.bg} text-white text-xs font-bold px-5 py-2.5 rounded-lg hover:opacity-90 transition-all shrink-0 disabled:opacity-50`}
        >
          {bundle.createMut.isPending ? "Adding..." : `+ Add ${config.itemLabel}`}
        </button>
      </form>
      {formError && <p className="text-[11px] font-bold text-red-500 -mt-3 px-1">{formError}</p>}

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full text-xs">
          <thead>
            <tr className="bg-gray-50 text-gray-400 text-[10px] font-black uppercase tracking-wider">
              <th className="text-left px-5 py-3">{config.itemLabel}</th>
              <th className="text-right px-5 py-3 w-40">Actions</th>
            </tr>
          </thead>
          <tbody>
            {bundle.isLoading && (
              <tr>
                <td colSpan={2} className="px-5 py-10 text-center text-gray-400 font-medium">
                  Loading...
                </td>
              </tr>
            )}
            {!bundle.isLoading && items.length === 0 && (
              <tr>
                <td colSpan={2} className="px-5 py-10 text-center text-gray-400 font-medium">
                  No items yet — add the first {config.itemLabel.toLowerCase()} above.
                </td>
              </tr>
            )}
            {items.map((item) => {
              const isEditing = editingId === item.id;
              return (
                <tr key={item.id} className="border-t border-gray-50 hover:bg-gray-50/60">
                  <td className="px-5 py-3 align-top">
                    {isEditing ? (
                      <>
                        <input
                          value={editName}
                          onChange={(e) => onEditNameChange(e.target.value)}
                          className="w-full bg-gray-50 rounded-lg p-2 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0"
                          autoFocus
                        />
                        {editError && <p className="text-[10px] font-bold text-red-500 mt-1">{editError}</p>}
                      </>
                    ) : (
                      <span className="font-bold text-gray-800">{item.label}</span>
                    )}
                  </td>
                  <td className="px-5 py-3 align-top">
                    <div className="flex justify-end gap-2">
                      {isEditing ? (
                        <>
                          <button
                            onClick={() => onSaveEdit(item.id)}
                            disabled={bundle.updateMut.isPending}
                            className="text-[11px] font-bold text-white bg-emerald-600 hover:bg-emerald-700 px-3 py-1.5 rounded-lg transition-all disabled:opacity-50"
                          >
                            {bundle.updateMut.isPending ? "Saving..." : "Save"}
                          </button>
                          <button
                            onClick={onCancelEdit}
                            className="text-[11px] font-bold text-gray-500 bg-gray-100 hover:bg-gray-200 px-3 py-1.5 rounded-lg transition-all"
                          >
                            Cancel
                          </button>
                        </>
                      ) : (
                        <>
                          <button
                            onClick={() => onStartEdit(item)}
                            className={`text-[11px] font-bold ${accent.text} ${accent.lightBg} hover:opacity-80 px-3 py-1.5 rounded-lg transition-all`}
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => onRequestDelete(item)}
                            className="text-[11px] font-bold text-red-500 bg-red-50 hover:bg-red-100 px-3 py-1.5 rounded-lg transition-all"
                          >
                            Delete
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </main>
  );
}
