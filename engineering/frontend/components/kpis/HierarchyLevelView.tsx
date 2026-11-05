"use client";

import { useState, type FormEvent } from "react";
import {
  HIERARCHIES,
  ACCENT_CLASSES,
  hierarchyLevelHref,
  type HierarchyId,
} from "@/lib/kpi-config";
import { getErrorMessage } from "@/lib/errors";
import { useFlashToast } from "@/hooks/use-flash-toast";
import type { HierarchyRow } from "@/types/kpi";
import type { HierarchyLevelBundle } from "@/types/kpi-admin";
import HierarchyLevelPanel from "@/components/kpis/HierarchyLevelPanel";
import ConfirmDeleteModal from "@/components/kpis/ConfirmDeleteModal";
import Toast from "@/components/kpis/Toast";

interface HierarchyLevelViewProps {
  hierarchyId: HierarchyId;
  levelIndex: number;
  /** Every level of this hierarchy, top -> leaf (needed for the parent
   * picker and the cascade-delete warning). */
  levels: HierarchyLevelBundle[];
  pageSize: number;
}

export default function HierarchyLevelView({
  hierarchyId,
  levelIndex,
  levels,
  pageSize,
}: HierarchyLevelViewProps) {
  const hierarchyDef = HIERARCHIES[hierarchyId];
  const hAccent = ACCENT_CLASSES[hierarchyDef.accent];
  const levelDef = hierarchyDef.levels[levelIndex];
  const parentLevelDef = levelIndex > 0 ? hierarchyDef.levels[levelIndex - 1] : null;
  const levelBundle = levels[levelIndex];
  const parentItems = levelIndex > 0 ? levels[levelIndex - 1].items : [];

  const [newName, setNewName] = useState("");
  const [newCode, setNewCode] = useState("");
  const [newParentId, setNewParentId] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editCode, setEditCode] = useState("");
  const [editParentId, setEditParentId] = useState("");
  const [editError, setEditError] = useState<string | null>(null);

  const [deleteTarget, setDeleteTarget] = useState<HierarchyRow | null>(null);
  const { toast, flashToast } = useFlashToast();

  // CREATE
  function handleAdd(e: FormEvent) {
    e.preventDefault();

    const trimmedName = newName.trim();
    const trimmedCode = newCode.trim();

    if (!trimmedName) {
      setFormError(`${levelDef.itemLabel} name is required.`);
      return;
    }
    if (!trimmedCode) {
      setFormError(`${levelDef.itemLabel} code is required.`);
      return;
    }
    if (parentLevelDef && !newParentId) {
      setFormError(`Select a parent ${parentLevelDef.itemLabel}.`);
      return;
    }
    setFormError(null);

    const payload: Record<string, unknown> = { name: trimmedName, code: trimmedCode };
    if (levelBundle.parentField) {
      payload[levelBundle.parentField] = Number(newParentId);
    }

    levelBundle.createMut.mutate(payload, {
      onSuccess: () => {
        setNewName("");
        setNewCode("");
        setNewParentId("");
        flashToast(`Added "${trimmedName}"`);
      },
      onError: (err) => setFormError(getErrorMessage(err, "Failed to create item.")),
    });
  }

  // UPDATE
  function startEdit(item: HierarchyRow) {
    setEditingId(item.id);
    setEditName(item.name);
    setEditCode(item.code);
    setEditParentId(item.parentId != null ? String(item.parentId) : "");
    setEditError(null);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditError(null);
  }

  function saveEdit(id: number) {
    const trimmedName = editName.trim();
    const trimmedCode = editCode.trim();

    if (!trimmedName) {
      setEditError(`${levelDef.itemLabel} name is required.`);
      return;
    }
    if (!trimmedCode) {
      setEditError(`${levelDef.itemLabel} code is required.`);
      return;
    }
    if (parentLevelDef && !editParentId) {
      setEditError(`Select a parent ${parentLevelDef.itemLabel}.`);
      return;
    }

    const payload: Record<string, unknown> = { name: trimmedName, code: trimmedCode };
    if (levelBundle.parentField) {
      payload[levelBundle.parentField] = Number(editParentId);
    }

    levelBundle.updateMut.mutate(
      { id, payload },
      {
        onSuccess: () => {
          setEditingId(null);
          setEditError(null);
          flashToast(`Updated "${trimmedName}"`);
        },
        onError: (err) => setEditError(getErrorMessage(err, "Failed to update item.")),
      }
    );
  }

  // DELETE
  // Note: the DB has ON DELETE CASCADE on every level's parent_id foreign
  // key, so deleting a row here removes all its descendants server-side
  // automatically — the delete hooks (see hooks/use-kpi.ts) already
  // invalidate every level's cache for this hierarchy as a result.
  function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    levelBundle.deleteMut.mutate(target.id, {
      onSuccess: () => {
        flashToast(`Deleted "${target.name}"`);
        setDeleteTarget(null);
      },
      onError: (err) => {
        flashToast(getErrorMessage(err, "Failed to delete item."));
        setDeleteTarget(null);
      },
    });
  }

  // Rows nested below this one are deleted with it. They can't be counted
  // reliably here: the last level is paginated, so only one page is loaded.
  const hasLowerLevels = levelIndex < levels.length - 1;
  const levelCounts = levels.map((lvl) => lvl.total ?? lvl.items.length);

  return (
    <>
      <HierarchyLevelPanel
        hierarchyDef={hierarchyDef}
        hAccent={hAccent}
        activeLevelIndex={levelIndex}
        levelCounts={levelCounts}
        levelHref={(index) => hierarchyLevelHref(hierarchyId, index)}
        pageSize={pageSize}
        levelBundle={levelBundle}
        parentItems={parentItems}
        newName={newName}
        onNewNameChange={setNewName}
        newCode={newCode}
        onNewCodeChange={setNewCode}
        newParentId={newParentId}
        onNewParentIdChange={setNewParentId}
        formError={formError}
        onSubmit={handleAdd}
        editingId={editingId}
        editName={editName}
        onEditNameChange={setEditName}
        editCode={editCode}
        onEditCodeChange={setEditCode}
        editParentId={editParentId}
        onEditParentIdChange={setEditParentId}
        editError={editError}
        onStartEdit={startEdit}
        onCancelEdit={cancelEdit}
        onSaveEdit={saveEdit}
        onRequestDelete={setDeleteTarget}
      />

      {deleteTarget && (
        <ConfirmDeleteModal
          title={`Delete ${levelDef.itemLabel}?`}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={handleDelete}
          isPending={levelBundle.deleteMut.isPending}
        >
          <p>
            This will permanently remove{" "}
            <span className="font-bold text-gray-800">
              &quot;{deleteTarget.code} — {deleteTarget.name}&quot;
            </span>
            .
          </p>
          {hasLowerLevels && (
            <p className="font-bold text-red-500 mt-2">
              It also permanently removes every item nested underneath it in the lower levels.
            </p>
          )}
        </ConfirmDeleteModal>
      )}

      <Toast message={toast} />
    </>
  );
}
