import type { KPICategoryConfig } from "@/lib/kpi-config";
import { CRUD_ACTIONS } from "@/lib/scopes";
import { THEME } from "@/lib/theme";
import RowAction from "@/components/kpis/RowAction";
import {
  PermissionsHint,
  usePermissionUi,
} from "@/components/kpis/permission-ui";
import type { FlatBundle, FlatItem } from "@/types/kpi-admin";

interface FlatCategoryPanelProps {
  config: KPICategoryConfig;
  bundle: FlatBundle;
  onAdd: () => void;
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
  bundle,
  onAdd,
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
  const { actionProps, hintProps, emptyMessage } = usePermissionUi(bundle.permissions, config.itemLabel);

  return (
    <main className="flex-1 p-4 mx-auto space-y-4">
      <div className="flex justify-between items-start gap-4">
        <div>
          <h2 className="text-lg font-black text-gray-900">{config.title}</h2>
          <p className="text-xs text-gray-400 mt-1">{config.description}</p>
        </div>
        <button
          type="button"
          onClick={onAdd}
          {...actionProps(CRUD_ACTIONS.create)}
          title={actionProps(CRUD_ACTIONS.create).title ?? `Add a new ${config.itemLabel.toLowerCase()}`}
          className={`cursor-pointer shrink-0 ${THEME.button} text-xs font-bold px-5 py-2.5 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
        >
          + Add {config.itemLabel}
        </button>
      </div>

      {bundle.isError && (
        <p className="text-xs font-bold text-red-500">
          Failed to load {config.title.toLowerCase()}. Please refresh.
        </p>
      )}

      <PermissionsHint {...hintProps} />

      <div className="bg-white rounded-md shadow-sm overflow-hidden">
        <table className="w-full text-xs">
          <thead>
            <tr className="bg-gray-50 text-gray-400 text-[10px] font-black uppercase tracking-wider">
              <th className="text-left px-5 py-3">{config.itemLabel}</th>
              <th className="text-right px-5 py-3 w-28">Actions</th>
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
                  {emptyMessage}
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
                    <div className="flex justify-end items-center gap-2">
                      {isEditing ? (
                        <>
                          <button
                            onClick={() => onSaveEdit(item.id)}
                            {...actionProps(CRUD_ACTIONS.update, bundle.updateMut.isPending)}
                            className={`cursor-pointer text-[11px] font-bold ${THEME.button} px-3 py-1.5 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
                          >
                            {bundle.updateMut.isPending ? "Saving..." : "Save"}
                          </button>
                          <button
                            onClick={onCancelEdit}
                            className={`cursor-pointer text-[11px] font-bold ${THEME.buttonSecondary} px-3 py-1.5 rounded-full transition-all`}
                          >
                            Cancel
                          </button>
                        </>
                      ) : (
                        <>
                          <RowAction
                            kind="edit"
                            label="Edit"
                            subject={item.label}
                            onClick={() => onStartEdit(item)}
                            {...actionProps(CRUD_ACTIONS.update)}
                          />
                          <RowAction
                            kind="delete"
                            label="Delete"
                            subject={item.label}
                            onClick={() => onRequestDelete(item)}
                            {...actionProps(CRUD_ACTIONS.delete)}
                          />
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
