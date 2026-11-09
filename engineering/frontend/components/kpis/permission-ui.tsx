"use client";

import { useId } from "react";
import {
  disabledReason,
  PERMISSION_STATUS,
  permissionsHintText,
  type ActionPermissions,
  type PermissionStatus,
} from "@/lib/action-permissions";
import { CRUD_ACTIONS, type CrudAction } from "@/lib/scopes";

/** Extra classes for a form control that permissions can disable. */
export const DISABLED_INPUT_CLASS = "disabled:cursor-not-allowed disabled:opacity-60";
/** Extra classes for a row's Edit/Delete buttons when disabled. */
export const DISABLED_ROW_ACTION_CLASS = "disabled:opacity-40 disabled:cursor-not-allowed";

/**
 * Everything a KPI panel needs to reflect permissions, shared by the flat and hierarchy panels so
 * both behave and explain the same way:
 * - `allowed`: which actions are allowed, keyed by action;
 * - `actionProps(action, isPending)`: `disabled`, tooltip and hint link for a button;
 * - `inputProps(action)`: `disabled` and hint link for a form control;
 * - `hintProps`: for `<PermissionsHint>`;
 * - `emptyMessage`: the "no items" text, inviting to add only when the user can.
 */
export function usePermissionUi(permissions: ActionPermissions, itemLabel: string) {
  const hintId = useId();
  const hint = permissionsHintText(permissions, itemLabel);
  const { allowed } = permissions;

  const describedBy = (action: CrudAction) => (!allowed[action] && hint ? hintId : undefined);

  const actionProps = (action: CrudAction, isPending = false) => ({
    disabled: !allowed[action] || isPending,
    title: disabledReason(permissions, action, itemLabel),
    "aria-describedby": describedBy(action),
  });

  const inputProps = (action: CrudAction) => ({
    disabled: !allowed[action],
    "aria-describedby": describedBy(action),
  });

  const emptyMessage = allowed[CRUD_ACTIONS.create]
    ? `No items yet — add the first ${itemLabel.toLowerCase()}.`
    : "No items yet.";

  return {
    allowed,
    actionProps,
    inputProps,
    hintProps: { id: hintId, hint, status: permissions.status },
    emptyMessage,
  };
}

/**
 * `disabled` and tooltip for one action on one resource, without a hook, for rows that span
 * several resources (a tree row acts on its own level and on the level below it).
 */
export function rowActionProps(permissions: ActionPermissions, action: CrudAction, itemLabel: string) {
  return {
    disabled: !permissions.allowed[action],
    title: disabledReason(permissions, action, itemLabel),
  };
}

/** The one-line explanation shown above a KPI table when some actions are disabled. */
export function PermissionsHint({
  id,
  hint,
  status,
}: {
  id: string;
  hint: string | null;
  status: PermissionStatus;
}) {
  if (!hint) return null;
  const tone = status === PERMISSION_STATUS.unavailable ? "text-red-600 bg-red-50" : "text-gray-500 bg-gray-100";
  return (
    // A live region, so screen readers announce when "Checking…" turns into the result.
    <p id={id} role="status" aria-live="polite" className={`text-xs font-semibold rounded-md px-3 py-2 ${tone}`}>
      {hint}
    </p>
  );
}
