"use client";

import { useState } from "react";
import { HIERARCHIES, type HierarchyId } from "@/lib/kpi-config";
import { getErrorMessage } from "@/lib/errors";
import { CRUD_ACTIONS } from "@/lib/scopes";
import { THEME } from "@/lib/theme";
import { useFlashToast } from "@/hooks/use-flash-toast";
import type { HierarchyRow } from "@/types/kpi";
import { useAdminHierarchyChildren } from "@/hooks/use-admin-hierarchy-children";
import { buildAncestors } from "@/lib/hierarchy-ancestors";
import type { HierarchyLevelMutations, HierarchyRoots, HierarchySelection } from "@/types/kpi-admin";
import HierarchyTreeNode, { nodeKey } from "@/components/kpis/HierarchyTreeNode";
import HierarchyDetailPanel from "@/components/kpis/HierarchyDetailPanel";
import CreateItemModal from "@/components/kpis/CreateItemModal";
import ConfirmDeleteModal from "@/components/kpis/ConfirmDeleteModal";
import { PermissionsHint, usePermissionUi } from "@/components/kpis/permission-ui";
import Toast from "@/components/kpis/Toast";

interface HierarchyTreeViewProps {
  hierarchyId: HierarchyId;
  /**
   * Every item of each level that can be a parent, top first: `lists[0]` is the top level,
   * which the tree starts from. The leaf level has no list. Feeds the "Parent" dropdown.
   */
  lists: HierarchyRoots[];
  /** The mutations and permissions of every level, top -> leaf. */
  levels: HierarchyLevelMutations[];
}

/** Where a new item is being created: its level, and the item it goes under (null for the top level). */
interface CreateTarget {
  levelIndex: number;
  parent: HierarchySelection | null;
}

// Both boxes take the height of the taller one, up to the viewport-based cap,
// and scroll on their own beyond it.
const PANEL_CLASS =
  "bg-white rounded-md shadow-sm flex flex-col min-h-48 max-h-[calc(100vh-14rem)] overflow-hidden";

export default function HierarchyTreeView({ hierarchyId, lists, levels }: HierarchyTreeViewProps) {
  const roots = lists[0];
  const hierarchyDef = HIERARCHIES[hierarchyId];
  const rootLevelDef = hierarchyDef.levels[0];
  const lastLevelIndex = hierarchyDef.levels.length - 1;

  const [pickedSelection, setSelection] = useState<HierarchySelection | null>(null);
  // Until something is picked (or after the picked item is deleted with no parent
  // to fall back to), the first top-level item is selected.
  const firstRoot = roots.items[0];
  const selection: HierarchySelection | null =
    pickedSelection ?? (firstRoot ? { levelIndex: 0, row: firstRoot, ancestors: [] } : null);
  const [expandedKeys, setExpandedKeys] = useState<ReadonlySet<string>>(new Set());
  const [createTarget, setCreateTarget] = useState<CreateTarget | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<HierarchySelection | null>(null);
  const { toast, flashToast } = useFlashToast();
  const rootPermissionUi = usePermissionUi(levels[0].permissions, rootLevelDef.itemLabel);

  const selectedKey = selection ? nodeKey(selection.levelIndex, selection.row.id) : null;

  function expand(key: string) {
    setExpandedKeys((prev) => new Set(prev).add(key));
  }

  function toggle(key: string) {
    setExpandedKeys((prev) => {
      const next = new Set(prev);
      if (!next.delete(key)) next.add(key);
      return next;
    });
  }

  function handleCreated(name: string) {
    if (createTarget?.parent) {
      expand(nodeKey(createTarget.parent.levelIndex, createTarget.parent.row.id));
    }
    setCreateTarget(null);
    flashToast(`Added "${name}"`);
  }

  // Saving can also move the item to another parent. It then shows up under the new parent,
  // so the tree opens that branch and the breadcrumb is rebuilt to match.
  function handleUpdated(previous: HierarchySelection, row: HierarchyRow) {
    const moved = row.parentId !== previous.row.parentId;
    let ancestors = previous.ancestors;
    if (moved && row.parentId != null) {
      const parentLevelIndex = previous.levelIndex - 1;
      const parent = lists[parentLevelIndex]?.items.find((item) => item.id === row.parentId);
      if (parent) {
        // The chain ends with the new parent, so opening every node in it shows the item.
        const chain = buildAncestors(lists, parentLevelIndex, parent);
        ancestors = chain;
        setExpandedKeys((prev) => {
          const next = new Set(prev);
          for (const node of chain) next.add(nodeKey(node.levelIndex, node.row.id));
          return next;
        });
      }
    }
    setSelection({ levelIndex: previous.levelIndex, row, ancestors });
    flashToast(`${moved ? "Moved" : "Updated"} "${row.name}"`);
  }

  // The backend soft-deletes the row together with all its descendants, and
  // the delete hooks invalidate every level's cache for this hierarchy.
  function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    const level = levels[target.levelIndex];
    // The bin icons are already disabled without permission; this is a backstop.
    if (!level.permissions.allowed[CRUD_ACTIONS.delete]) return setDeleteTarget(null);

    const targetKey = nodeKey(target.levelIndex, target.row.id);
    level.deleteMut.mutate(target.row.id, {
      onSuccess: () => {
        flashToast(`Deleted "${target.row.name}"`);
        // Drop the selection if it was the deleted item or sat under it.
        const wasRemoved =
          selectedKey === targetKey ||
          selection?.ancestors.some((a) => nodeKey(a.levelIndex, a.row.id) === targetKey);
        if (wasRemoved) setSelection(target.ancestors.at(-1) ?? null);
        setDeleteTarget(null);
      },
      onError: (err) => {
        flashToast(getErrorMessage(err, "Failed to delete item."));
        setDeleteTarget(null);
      },
    });
  }

  const createLevelDef = createTarget ? hierarchyDef.levels[createTarget.levelIndex] : null;
  const createParentLevelDef = createTarget?.parent ? hierarchyDef.levels[createTarget.parent.levelIndex] : null;

  return (
    <main className="flex-1 p-4 mx-auto w-full space-y-4">
      <div>
        <h2 className="text-lg font-black text-gray-900">{hierarchyDef.label} Hierarchy</h2>
        <p className="text-xs text-gray-400 mt-1">
          Expand an item to see what is under it,
          select it to edit, and use + and the bin icon on a row to add under it or delete it.
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
        <section className={PANEL_CLASS}>
          <div className="flex items-center justify-between gap-2 p-3 border-b border-gray-100 shrink-0">
            <h3 className="text-[10px] font-black uppercase text-gray-400 tracking-wider">
              {rootLevelDef.title}s
            </h3>
            <button
              type="button"
              onClick={() => setCreateTarget({ levelIndex: 0, parent: null })}
              {...rootPermissionUi.actionProps(CRUD_ACTIONS.create)}
              title={rootPermissionUi.actionProps(CRUD_ACTIONS.create).title ?? `Add a new ${rootLevelDef.itemLabel.toLowerCase()}`}
              className={`cursor-pointer ${THEME.button} text-[11px] font-bold px-3 py-1.5 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
            >
              + Add {rootLevelDef.itemLabel}
            </button>
          </div>

          {roots.isError && (
            <p className="text-xs font-bold text-red-500 px-3 pt-3">
              Failed to load {rootLevelDef.title.toLowerCase()} data. Please refresh.
            </p>
          )}
          <div className="px-3 pt-3 empty:hidden">
            <PermissionsHint {...rootPermissionUi.hintProps} />
          </div>
          {roots.isLoading && <p className="px-2 py-6 text-center text-xs text-gray-400">Loading...</p>}
          {!roots.isLoading && !roots.isError && roots.items.length === 0 && (
            <p className="px-2 py-6 text-center text-xs text-gray-400">{rootPermissionUi.emptyMessage}</p>
          )}
          <ul className="flex-1 min-h-0 overflow-y-auto p-3">
            {roots.items.map((row) => (
              <HierarchyTreeNode
                key={row.id}
                hierarchyId={hierarchyId}
                levels={levels}
                expandedKeys={expandedKeys}
                selectedKey={selectedKey}
                onToggle={toggle}
                onSelect={setSelection}
                onAddChild={(parent) => setCreateTarget({ levelIndex: parent.levelIndex + 1, parent })}
                onDelete={setDeleteTarget}
                levelIndex={0}
                row={row}
                ancestors={[]}
              />
            ))}
          </ul>
        </section>

        <section className={`${PANEL_CLASS} p-5`}>
          <div className="flex-1 min-h-0 overflow-y-auto">
            {selection ? (
              <HierarchyDetailPanel
                hierarchyId={hierarchyId}
                selection={selection}
                mutations={levels[selection.levelIndex]}
                showPermissionsHint={selection.levelIndex > 0}
                onSelectAncestor={setSelection}
                parentList={selection.levelIndex > 0 ? lists[selection.levelIndex - 1] : null}
                parentLabel={selection.levelIndex > 0 ? hierarchyDef.levels[selection.levelIndex - 1].itemLabel : null}
                onUpdated={(row) => handleUpdated(selection, row)}
              />
            ) : (
              <p className="text-xs text-gray-400 text-center py-10">
                Select an item in the tree to view or edit it.
              </p>
            )}
          </div>
        </section>
      </div>

      {createTarget && createLevelDef && (
        <CreateItemModal
          itemLabel={createLevelDef.itemLabel}
          parent={createTarget.parent?.row ?? null}
          parentLabel={createParentLevelDef?.itemLabel ?? null}
          mutations={levels[createTarget.levelIndex]}
          onCancel={() => setCreateTarget(null)}
          onCreated={handleCreated}
        />
      )}

      {deleteTarget && (
        <ConfirmDeleteModal
          title={`Delete ${hierarchyDef.levels[deleteTarget.levelIndex].itemLabel}?`}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={handleDelete}
          isPending={levels[deleteTarget.levelIndex].deleteMut.isPending}
        >
          <p>
            This will permanently remove{" "}
            <span className="font-bold text-gray-800">
              &quot;{deleteTarget.row.code} — {deleteTarget.row.name}&quot;
            </span>
            .
          </p>
          {deleteTarget.levelIndex < lastLevelIndex && (
            <DescendantWarning hierarchyId={hierarchyId} selection={deleteTarget} />
          )}
        </ConfirmDeleteModal>
      )}

      <Toast message={toast} />
    </main>
  );
}

/** Only the direct children are counted; deeper levels are removed too but aren't loaded here. */
function DescendantWarning({ hierarchyId, selection }: { hierarchyId: HierarchyId; selection: HierarchySelection }) {
  const childrenQ = useAdminHierarchyChildren(hierarchyId, selection.levelIndex, selection.row.id);
  const childLabel = HIERARCHIES[hierarchyId].levels[selection.levelIndex + 1].itemLabel.toLowerCase();
  const count = childrenQ.data?.length;

  return (
    <p className="font-bold text-red-500 mt-2">
      {count ? `It has ${count} direct ${childLabel}${count === 1 ? "" : "s"}. ` : ""}
      It also permanently removes every item nested underneath it in the lower levels.
    </p>
  );
}
