"use client";

import { useState, type FormEvent } from "react";
import { KPI_CATEGORIES, ACCENT_CLASSES, type KPICategoryKey } from "@/lib/kpi-config";
import { getErrorMessage } from "@/lib/errors";
import { permissionGuards } from "@/lib/action-permissions";
import { CRUD_ACTIONS } from "@/lib/scopes";
import { useFlashToast } from "@/hooks/use-flash-toast";
import type { FlatBundle, FlatItem } from "@/types/kpi-admin";
import FlatCategoryPanel from "@/components/kpis/FlatCategoryPanel";
import ConfirmDeleteModal from "@/components/kpis/ConfirmDeleteModal";
import Toast from "@/components/kpis/Toast";

interface FlatCategoryViewProps {
  category: KPICategoryKey;
  bundle: FlatBundle;
}

export default function FlatCategoryView({ category, bundle }: FlatCategoryViewProps) {
  const config = KPI_CATEGORIES[category];
  const accent = ACCENT_CLASSES[config.accent];

  const [newName, setNewName] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editError, setEditError] = useState<string | null>(null);

  const [deleteTarget, setDeleteTarget] = useState<FlatItem | null>(null);
  const { toast, flashToast } = useFlashToast();
  const guards = permissionGuards<FlatItem>(bundle.permissions, config.itemLabel, { setEditError, setDeleteTarget, flashToast });

  // CREATE
  function handleAdd(e: FormEvent) {
    e.preventDefault();
    if (!guards.canStart(CRUD_ACTIONS.create)) return;
    const trimmed = newName.trim();
    if (!trimmed) {
      setFormError(`${config.itemLabel} name is required.`);
      return;
    }
    setFormError(null);
    bundle.createMut.mutate(
      { [bundle.labelField]: trimmed },
      {
        onSuccess: () => {
          setNewName("");
          flashToast(`Added "${trimmed}"`);
        },
        onError: (err) => setFormError(getErrorMessage(err, "Failed to create item.")),
      }
    );
  }

  // UPDATE
  function startEdit(item: FlatItem) {
    if (!guards.canStart(CRUD_ACTIONS.update)) return;
    setEditingId(item.id);
    setEditName(item.label);
    setEditError(null);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditError(null);
  }

  function saveEdit(id: number) {
    if (!guards.canSave()) return;
    const trimmed = editName.trim();
    if (!trimmed) {
      setEditError(`${config.itemLabel} name is required.`);
      return;
    }
    bundle.updateMut.mutate(
      { id, payload: { [bundle.labelField]: trimmed } },
      {
        onSuccess: () => {
          setEditingId(null);
          setEditError(null);
          flashToast(`Updated "${trimmed}"`);
        },
        onError: (err) => setEditError(getErrorMessage(err, "Failed to update item.")),
      }
    );
  }

  // DELETE
  function handleDelete() {
    if (!deleteTarget) return;
    if (!guards.canConfirmDelete()) return;
    const target = deleteTarget;
    bundle.deleteMut.mutate(target.id, {
      onSuccess: () => {
        flashToast(`Deleted "${target.label}"`);
        setDeleteTarget(null);
      },
      onError: (err) => {
        flashToast(getErrorMessage(err, "Failed to delete item."));
        setDeleteTarget(null);
      },
    });
  }

  return (
    <>
      <FlatCategoryPanel
        config={config}
        accent={accent}
        bundle={bundle}
        newName={newName}
        onNewNameChange={setNewName}
        formError={formError}
        onSubmit={handleAdd}
        editingId={editingId}
        editName={editName}
        onEditNameChange={setEditName}
        editError={editError}
        onStartEdit={startEdit}
        onCancelEdit={cancelEdit}
        onSaveEdit={saveEdit}
        onRequestDelete={guards.requestDelete}
      />

      {deleteTarget && (
        <ConfirmDeleteModal
          title={`Delete ${config.itemLabel}?`}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={handleDelete}
          isPending={bundle.deleteMut.isPending}
        >
          This will permanently remove{" "}
          <span className="font-bold text-gray-800">&quot;{deleteTarget.label}&quot;</span> from{" "}
          {config.title}. This action can&apos;t be undone.
        </ConfirmDeleteModal>
      )}

      <Toast message={toast} />
    </>
  );
}
