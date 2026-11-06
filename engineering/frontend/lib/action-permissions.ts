import { CRUD_ACTION_LABELS, CRUD_ACTIONS, scopeFor, type CrudAction, type ResourceScope } from "@/lib/scopes";

/** Where the signed-in user's granted permissions stand. */
export const PERMISSION_STATUS = {
  /** The granted permissions haven't arrived yet. */
  loading: "loading",
  /** They arrived (possibly granting nothing for a given resource). */
  ready: "ready",
  /** They couldn't be read, so we can't tell what the user may do. */
  unavailable: "unavailable",
} as const;

export type PermissionStatus = (typeof PERMISSION_STATUS)[keyof typeof PERMISSION_STATUS];

/** The signed-in user's granted permissions, as shared across the KPI pages. */
export interface GrantedPermissions {
  status: PermissionStatus;
  can: (permission: string) => boolean;
}

/** Which CRUD actions the signed-in user may perform on a resource. Drives the UI only. */
export interface ActionPermissions {
  status: PermissionStatus;
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;
}

/**
 * Works out the permission status from the SDK's `useGrantedPermissions()` result.
 *
 * The SDK returns an empty list both when the user has no session and when reading the token
 * fails. A signed-in user's token always carries at least `openid`, so an empty list for a
 * signed-in user means the permissions couldn't be read.
 */
export function resolvePermissionStatus(
  isSignedIn: boolean,
  isLoading: boolean,
  permissions: readonly string[],
): PermissionStatus {
  if (isLoading) return PERMISSION_STATUS.loading;
  if (isSignedIn && permissions.length === 0) return PERMISSION_STATUS.unavailable;
  return PERMISSION_STATUS.ready;
}

/**
 * Maps a resource's scope (e.g. "experiences") to the `<scope>:create|update|delete` permissions.
 * Fails closed: unless permissions are `ready`, nothing is allowed.
 *
 * The backend enforces the same scopes; this only decides what the UI enables.
 */
export function resolveActionPermissions(scope: ResourceScope, granted: GrantedPermissions): ActionPermissions {
  const ready = granted.status === PERMISSION_STATUS.ready;
  return {
    status: granted.status,
    canCreate: ready && granted.can(scopeFor(scope, CRUD_ACTIONS.create)),
    canUpdate: ready && granted.can(scopeFor(scope, CRUD_ACTIONS.update)),
    canDelete: ready && granted.can(scopeFor(scope, CRUD_ACTIONS.delete)),
  };
}

/** Every user-facing permission message, in one place so the wording stays consistent. */
export const PERMISSION_MESSAGES = {
  checking: "Checking your permissions…",
  unavailable: "Couldn't load your permissions. Refresh the page to try again.",
} as const;

/** "You don't have permission to add, edit or delete experience items." */
export function permissionDeniedMessage(actions: readonly CrudAction[], itemLabel: string): string {
  const words = actions.map((action) => CRUD_ACTION_LABELS[action]);
  const list = words.length === 1 ? words[0] : `${words.slice(0, -1).join(", ")} or ${words.at(-1)}`;
  return `You don't have permission to ${list} ${itemLabel.toLowerCase()} items.`;
}

/** Why an action is disabled, for its tooltip. Undefined when the action is allowed. */
export function disabledReason(
  permissions: ActionPermissions,
  allowed: boolean,
  action: CrudAction,
  itemLabel: string,
): string | undefined {
  if (allowed) return undefined;
  if (permissions.status === PERMISSION_STATUS.loading) return PERMISSION_MESSAGES.checking;
  if (permissions.status === PERMISSION_STATUS.unavailable) return PERMISSION_MESSAGES.unavailable;
  return permissionDeniedMessage([action], itemLabel);
}

/**
 * A visible, one-line explanation of what the user can't do, or null when nothing needs
 * explaining (everything allowed, or permissions still loading).
 */
export function permissionsHint(permissions: ActionPermissions, itemLabel: string): string | null {
  if (permissions.status === PERMISSION_STATUS.loading) return null;
  if (permissions.status === PERMISSION_STATUS.unavailable) return PERMISSION_MESSAGES.unavailable;
  const missing = [
    !permissions.canCreate && CRUD_ACTIONS.create,
    !permissions.canUpdate && CRUD_ACTIONS.update,
    !permissions.canDelete && CRUD_ACTIONS.delete,
  ].filter((action): action is CrudAction => Boolean(action));
  return missing.length === 0 ? null : permissionDeniedMessage(missing, itemLabel);
}
