"use client";

import { useMemo, useState, type KeyboardEvent } from "react";
import { LockClosedIcon } from "@heroicons/react/16/solid";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
import {
  useCreateSource,
  useDeleteSource,
  useRenameSource,
  useSources,
  useUpdateSourceStatus,
} from "@/hooks/use-sources";
import { useActionPermissions } from "@/hooks/use-action-permissions";
import { useFlashToast } from "@/hooks/use-flash-toast";
import { PermissionsHint, usePermissionUi } from "@/components/kpis/permission-ui";
import RowAction from "@/components/kpis/RowAction";
import Toast from "@/components/kpis/Toast";
import AddSourceModal from "@/components/sources/AddSourceModal";
import DeleteSourceModal from "@/components/sources/DeleteSourceModal";
import DisableSourceModal from "@/components/sources/DisableSourceModal";
import SourceStatusSwitch from "@/components/sources/SourceStatusSwitch";
import { getErrorMessage } from "@/lib/errors";
import { CRUD_ACTIONS, RESOURCE_SCOPES } from "@/lib/scopes";
import { THEME } from "@/lib/theme";
import { SOURCE_TYPE_LABELS, SOURCE_TYPES, type Source, type SourceType } from "@/types/source";

const ITEM_LABEL = "Source";

/** A type tab, or "all" (a UI-only option, not a source type). */
type TypeFilter = "all" | SourceType;

const FILTER_TABS: { key: TypeFilter; label: string }[] = [
  { key: "all", label: "All" },
  ...SOURCE_TYPES.map((type) => ({ key: type, label: SOURCE_TYPE_LABELS[type] })),
];

const LINKED_NAME_HINT =
  "Linked to a data collector. The name must match exactly (case-sensitive), or new jobs from it are rejected.";
const AUTO_RENAME_HINT = "Automatic sources can't be renamed: the name must match its data collector exactly.";

/** Automatic sources first, then by name. */
function bySourceOrder(a: Source, b: Source) {
  if (a.type !== b.type) return a.type === "auto" ? -1 : 1;
  return a.name.localeCompare(b.name);
}

export default function SourcesPanel() {
  const sourceList = useSources();
  const createMut = useCreateSource();
  const renameMut = useRenameSource();
  const statusMut = useUpdateSourceStatus();
  const deleteMut = useDeleteSource();
  const permissions = useActionPermissions(RESOURCE_SCOPES.sources);
  const { allowed, actionProps, hintProps, emptyMessage } = usePermissionUi(permissions, ITEM_LABEL);
  const { toast, flashToast } = useFlashToast();

  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all");
  const [search, setSearch] = useState("");
  const [showAdd, setShowAdd] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editError, setEditError] = useState<string | null>(null);
  const [disableTarget, setDisableTarget] = useState<Source | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Source | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const all = useMemo(() => [...(sourceList.data?.sources ?? [])].sort(bySourceOrder), [sourceList.data]);

  const counts = useMemo(() => {
    const byType: Record<TypeFilter, number> = { all: all.length, auto: 0, manual: 0 };
    for (const s of all) byType[s.type]++;
    return byType;
  }, [all]);

  const visible = useMemo(() => {
    const needle = search.trim().toLowerCase();
    return all.filter(
      (s) => (typeFilter === "all" || s.type === typeFilter) && (!needle || s.name.toLowerCase().includes(needle)),
    );
  }, [all, typeFilter, search]);

  const pendingStatusId = statusMut.isPending ? statusMut.variables?.id : undefined;

  // status
  function setStatus(source: Source, isActive: boolean) {
    statusMut.mutate(
      { id: source.id, isActive },
      {
        onSuccess: () => {
          setDisableTarget(null);
          flashToast(`"${source.name}" ${isActive ? "enabled" : "disabled"}`);
        },
        onError: (err) => {
          setDisableTarget(null);
          flashToast(getErrorMessage(err, `Couldn't update "${source.name}".`));
        },
      },
    );
  }

  function onToggle(source: Source, next: boolean) {
    if (!allowed[CRUD_ACTIONS.update]) return;
    if (next) setStatus(source, true);
    else setDisableTarget(source);
  }

  // rename
  function startEdit(source: Source) {
    if (!allowed[CRUD_ACTIONS.update] || source.type !== "manual") return;
    setEditingId(source.id);
    setEditName(source.name);
    setEditError(null);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditError(null);
  }

  function saveEdit(source: Source) {
    const name = editName.trim();
    if (!name) return setEditError("Name is required.");
    if (name === source.name) return cancelEdit();
    renameMut.mutate(
      { id: source.id, name },
      {
        onSuccess: () => {
          cancelEdit();
          flashToast(`Renamed to "${name}"`);
        },
        onError: (err) => {
          const message = getErrorMessage(err, "Failed to rename source.");
          setEditError(
            message.includes("already exists")
              ? "A source with this name already exists. Deleted sources keep their names: add it again to restore it."
              : message,
          );
        },
      },
    );
  }

  function onEditKey(e: KeyboardEvent<HTMLInputElement>, source: Source) {
    if (e.key === "Enter") saveEdit(source);
    if (e.key === "Escape") cancelEdit();
  }

  // delete
  function requestDelete(source: Source) {
    if (!allowed[CRUD_ACTIONS.delete]) return;
    setDeleteError(null);
    setDeleteTarget(source);
  }

  function confirmDelete() {
    if (!deleteTarget) return;
    const { id, name } = deleteTarget;
    deleteMut.mutate(id, {
      onSuccess: () => {
        setDeleteTarget(null);
        if (editingId === id) cancelEdit();
        flashToast(`"${name}" deleted`);
      },
      onError: (err) => setDeleteError(getErrorMessage(err, "Failed to delete source.")),
    });
  }

  const createProps = actionProps(CRUD_ACTIONS.create);

  return (
    <main className="flex-1 p-4 mx-auto space-y-4">
      <div className="flex justify-between items-start gap-4">
        <div>
          <p className="text-xs text-gray-400">
            Where job vacancies come from: job websites collected automatically, and hand-entered sources such as
            newspapers.
          </p>
        </div>
        <button
          type="button"
          onClick={() => setShowAdd(true)}
          {...createProps}
          title={createProps.title ?? "Add a new source"}
          className={`cursor-pointer shrink-0 ${THEME.button} text-xs font-bold px-5 py-2.5 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
        >
          + Add Source
        </button>
      </div>

      <p className="flex items-start gap-2 text-xs text-blue-800 bg-blue-50 rounded-md px-3 py-2">
        <InformationCircleIcon className="w-4 h-4 shrink-0 mt-px" />
        <span>
          Changes take effect from the next collection run. Collection schedules are managed in the deployment, not
          here.
        </span>
      </p>

      {sourceList.isError && (
        <p className="text-xs font-bold text-red-500">
          {getErrorMessage(sourceList.error, "Failed to load sources.")} Please refresh.
        </p>
      )}

      <PermissionsHint {...hintProps} />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div role="tablist" aria-label="Filter by type" className="inline-flex rounded-full bg-gray-100 p-1">
          {FILTER_TABS.map((tab) => {
            const active = typeFilter === tab.key;
            return (
              <button
                key={tab.key}
                type="button"
                role="tab"
                aria-selected={active}
                onClick={() => setTypeFilter(tab.key)}
                className={`cursor-pointer text-[11px] font-bold px-3 py-1.5 rounded-full transition-all ${active ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-800"
                  }`}
              >
                {tab.label} <span className="text-gray-400 font-semibold">{counts[tab.key]}</span>
              </button>
            );
          })}
        </div>
        <input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search sources…"
          aria-label="Search sources"
          className="w-56 bg-white rounded-full px-4 py-2 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0"
        />
      </div>

      <div className="bg-white rounded-md shadow-sm overflow-hidden">
        <table className="w-full text-xs">
          <thead>
            <tr className="bg-gray-50 text-gray-400 text-[10px] font-black uppercase tracking-wider">
              <th className="text-left px-5 py-3">Name</th>
              <th className="text-left px-5 py-3 w-28">Type</th>
              <th className="text-left px-5 py-3 w-36">Status</th>
              <th className="text-right px-5 py-3 w-28">Actions</th>
            </tr>
          </thead>
          <tbody>
            {sourceList.isLoading && (
              <tr>
                <td colSpan={4} className="px-5 py-10 text-center text-gray-400 font-medium">
                  Loading...
                </td>
              </tr>
            )}
            {!sourceList.isLoading && !sourceList.isError && visible.length === 0 && (
              <tr>
                <td colSpan={4} className="px-5 py-10 text-center text-gray-400 font-medium">
                  {all.length === 0 ? emptyMessage : "No sources match your filters."}
                </td>
              </tr>
            )}
            {visible.map((source) => {
              const isAuto = source.type === "auto";
              const isEditing = editingId === source.id;
              const updateProps = actionProps(CRUD_ACTIONS.update);
              const deleteProps = actionProps(CRUD_ACTIONS.delete);
              const renameDisabled = updateProps.disabled || isAuto;
              const statusPending = pendingStatusId === source.id;

              return (
                <tr key={source.id} className="border-t border-gray-50 hover:bg-gray-50/60">
                  <td className="px-5 py-3 align-middle">
                    {isEditing ? (
                      <>
                        <input
                          value={editName}
                          onChange={(e) => setEditName(e.target.value)}
                          onKeyDown={(e) => onEditKey(e, source)}
                          aria-label={`New name for ${source.name}`}
                          className="w-full bg-gray-50 rounded-lg p-2 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0"
                          autoFocus
                        />
                        {editError && <p className="text-[10px] font-bold text-red-500 mt-1">{editError}</p>}
                      </>
                    ) : (
                      <div className={`flex items-center gap-2 ${source.is_active ? "" : "opacity-50"}`}>
                        <span className={`font-bold text-gray-800 ${isAuto ? "font-mono" : ""}`}>{source.name}</span>
                        {isAuto && (
                          <span
                            title={LINKED_NAME_HINT}
                            className="inline-flex items-center text-yellow-700 bg-yellow-50 border border-yellow-300 rounded-full px-1.5 py-0.5 cursor-help"
                          >
                            <LockClosedIcon className="w-3 h-3" />
                            <span className="sr-only">Linked name. {LINKED_NAME_HINT}</span>
                          </span>
                        )}
                      </div>
                    )}
                  </td>
                  <td className="px-5 py-3 align-middle">
                    <TypeBadge type={source.type} />
                  </td>
                  <td className="px-5 py-3 align-middle">
                    <div className="flex items-center gap-2">
                      <SourceStatusSwitch
                        checked={source.is_active}
                        label={source.name}
                        onChange={(next) => onToggle(source, next)}
                        disabled={updateProps.disabled || statusPending}
                        title={updateProps.title}
                        aria-describedby={updateProps["aria-describedby"]}
                      />
                      <span className={`text-[11px] font-bold ${source.is_active ? "text-emerald-700" : "text-gray-400"}`}>
                        {statusPending ? "Saving…" : source.is_active ? "Active" : "Inactive"}
                      </span>
                    </div>
                  </td>
                  <td className="px-5 py-3 align-middle">
                    <div className="flex justify-end items-center gap-2">
                      {isEditing ? (
                        <>
                          <button
                            type="button"
                            onClick={() => saveEdit(source)}
                            disabled={renameMut.isPending}
                            className={`cursor-pointer text-[11px] font-bold ${THEME.button} px-3 py-1.5 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
                          >
                            {renameMut.isPending ? "Saving..." : "Save"}
                          </button>
                          <button
                            type="button"
                            onClick={cancelEdit}
                            className={`cursor-pointer text-[11px] font-bold ${THEME.buttonSecondary} px-3 py-1.5 rounded-full transition-all`}
                          >
                            Cancel
                          </button>
                        </>
                      ) : (
                        <>
                          <RowAction
                            kind="edit"
                            label="Update"
                            subject={source.name}
                            onClick={() => startEdit(source)}
                            disabled={renameDisabled}
                            title={updateProps.title ?? (isAuto ? AUTO_RENAME_HINT : undefined)}
                            aria-describedby={updateProps["aria-describedby"]}
                          />
                          <RowAction
                            kind="delete"
                            label="Delete"
                            subject={source.name}
                            onClick={() => requestDelete(source)}
                            {...deleteProps}
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

      {showAdd && (
        <AddSourceModal
          existing={all}
          create={(payload, callbacks) => createMut.mutate(payload, callbacks)}
          isPending={createMut.isPending}
          onCancel={() => setShowAdd(false)}
          onCreated={(source) => {
            setShowAdd(false);
            flashToast(`"${source.name}" added`);
          }}
        />
      )}

      {disableTarget && (
        <DisableSourceModal
          source={disableTarget}
          isPending={statusMut.isPending}
          onCancel={() => setDisableTarget(null)}
          onConfirm={() => setStatus(disableTarget, false)}
        />
      )}

      {deleteTarget && (
        <DeleteSourceModal
          key={deleteTarget.id}
          source={deleteTarget}
          isPending={deleteMut.isPending}
          error={deleteError}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={confirmDelete}
        />
      )}

      <Toast message={toast} />
    </main>
  );
}

const TYPE_BADGE_COLORS: Record<SourceType, string> = {
  auto: "text-blue-700 bg-blue-50",
  manual: "text-purple-700 bg-purple-50",
};

function TypeBadge({ type }: { type: SourceType }) {
  return (
    <span className={`inline-flex text-[10px] font-black uppercase tracking-wider rounded-full px-2 py-0.5 ${TYPE_BADGE_COLORS[type]}`}>
      {SOURCE_TYPE_LABELS[type]}
    </span>
  );
}

