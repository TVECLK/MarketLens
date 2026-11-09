"use client";

import { useAdminHierarchyChildren } from "@/hooks/use-admin-hierarchy-children";
import { HIERARCHIES, type HierarchyId } from "@/lib/kpi-config";
import { CRUD_ACTIONS } from "@/lib/scopes";
import { THEME } from "@/lib/theme";
import RowAction from "@/components/kpis/RowAction";
import { rowActionProps } from "@/components/kpis/permission-ui";
import type { HierarchyLevelMutations, HierarchySelection } from "@/types/kpi-admin";
import type { HierarchyRow } from "@/types/kpi";

export function nodeKey(levelIndex: number, id: number): string {
  return `${levelIndex}:${id}`;
}

interface TreeContext {
  hierarchyId: HierarchyId;
  /** The mutations and permissions of every level, top -> leaf. */
  levels: HierarchyLevelMutations[];
  expandedKeys: ReadonlySet<string>;
  selectedKey: string | null;
  onToggle: (key: string) => void;
  onSelect: (selection: HierarchySelection) => void;
  onAddChild: (parent: HierarchySelection) => void;
  onDelete: (target: HierarchySelection) => void;
}

interface HierarchyTreeNodeProps extends TreeContext {
  levelIndex: number;
  row: HierarchyRow;
  ancestors: HierarchySelection[];
}

export default function HierarchyTreeNode(props: HierarchyTreeNodeProps) {
  const {
    hierarchyId, levels, expandedKeys, selectedKey, onToggle, onSelect, onAddChild, onDelete,
    levelIndex, row, ancestors,
  } = props;
  const levelDefs = HIERARCHIES[hierarchyId].levels;
  const key = nodeKey(levelIndex, row.id);
  const isLeaf = levelIndex === levelDefs.length - 1;
  const isExpanded = !isLeaf && expandedKeys.has(key);
  const isSelected = selectedKey === key;
  const self: HierarchySelection = { levelIndex, row, ancestors };

  return (
    <li>
      <div className={`flex items-center gap-1 rounded-md pr-2 ${isSelected ? THEME.selectedRow : "hover:bg-gray-50"}`}>
        {isLeaf ? (
          <span className="w-6 shrink-0" />
        ) : (
          <button
            type="button"
            onClick={() => onToggle(key)}
            aria-expanded={isExpanded}
            aria-label={`${isExpanded ? "Collapse" : "Expand"} ${row.name}`}
            className="cursor-pointer w-6 h-7 shrink-0 flex items-center justify-center text-gray-400 hover:text-gray-700"
          >
            <svg
              className={`w-3 h-3 transition-transform ${isExpanded ? "rotate-90" : ""}`}
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="3"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              <polyline points="9 6 15 12 9 18" />
            </svg>
          </button>
        )}
        <button
          type="button"
          onClick={() => onSelect(self)}
          aria-current={isSelected ? "true" : undefined}
          className="cursor-pointer flex-1 min-w-0 py-1.5 text-left"
        >
          <span className={`block text-xs truncate ${isSelected ? `font-black ${THEME.selectedText}` : "font-semibold text-gray-800"}`}>
            {row.name}
          </span>
        </button>
        {!isLeaf && (
          <RowAction
            kind="add"
            label={`Add ${levelDefs[levelIndex + 1].itemLabel}`}
            subject={`under ${row.name}`}
            onClick={() => onAddChild(self)}
            {...rowActionProps(levels[levelIndex + 1].permissions, CRUD_ACTIONS.create, levelDefs[levelIndex + 1].itemLabel)}
          />
        )}
        <RowAction
          kind="delete"
          label="Delete"
          subject={row.name}
          onClick={() => onDelete(self)}
          {...rowActionProps(levels[levelIndex].permissions, CRUD_ACTIONS.delete, levelDefs[levelIndex].itemLabel)}
        />
      </div>
      {isExpanded && <TreeChildren {...props} parent={self} />}
    </li>
  );
}

/** Mounted only while its parent is expanded, so a branch loads on first expand. */
function TreeChildren({ parent, ...ctx }: TreeContext & { parent: HierarchySelection }) {
  const childrenQ = useAdminHierarchyChildren(ctx.hierarchyId, parent.levelIndex, parent.row.id);
  const ancestors = [...parent.ancestors, parent];

  return (
    <ul className="ml-3 pl-2 border-l border-gray-100">
      {childrenQ.isLoading && <li className="px-2 py-1.5 text-[11px] text-gray-400">Loading...</li>}
      {childrenQ.isError && <li className="px-2 py-1.5 text-[11px] font-bold text-red-500">Failed to load.</li>}
      {childrenQ.data?.length === 0 && <li className="px-2 py-1.5 text-[11px] text-gray-400">No items.</li>}
      {childrenQ.data?.map((child) => (
        <HierarchyTreeNode
          key={child.id}
          {...ctx}
          levelIndex={parent.levelIndex + 1}
          row={child}
          ancestors={ancestors}
        />
      ))}
    </ul>
  );
}
