import { describe, expect, it, vi } from "vitest";
import {
  disabledReason,
  PERMISSION_MESSAGES,
  PERMISSION_STATUS,
  permissionDeniedMessage,
  permissionGuards,
  permissionsHint,
  resolveActionPermissions,
  resolvePermissionStatus,
  type ActionPermissions,
  type GrantedPermissions,
  type PermissionStatus,
} from "@/lib/action-permissions";
import { CRUD_ACTIONS, type CrudAction } from "@/lib/scopes";

const ready = (...granted: string[]): GrantedPermissions => ({
  status: PERMISSION_STATUS.ready,
  can: (permission) => granted.includes(permission),
});

/** Permissions with everything allowed except the listed actions. */
const allowedExcept = (denied: CrudAction[] = [], status: PermissionStatus = PERMISSION_STATUS.ready): ActionPermissions => ({
  status,
  allowed: {
    create: !denied.includes(CRUD_ACTIONS.create),
    update: !denied.includes(CRUD_ACTIONS.update),
    delete: !denied.includes(CRUD_ACTIONS.delete),
  },
});

const ALL = [CRUD_ACTIONS.create, CRUD_ACTIONS.update, CRUD_ACTIONS.delete];

describe("resolvePermissionStatus", () => {
  it("is loading while the SDK is loading", () => {
    expect(resolvePermissionStatus(true, true, [])).toBe(PERMISSION_STATUS.loading);
  });

  it("is unavailable when a signed-in user gets no permissions at all", () => {
    expect(resolvePermissionStatus(true, false, [])).toBe(PERMISSION_STATUS.unavailable);
  });

  it("is ready once permissions arrive", () => {
    expect(resolvePermissionStatus(true, false, ["openid", "profile"])).toBe(PERMISSION_STATUS.ready);
  });

  it("is ready (granting nothing) when signed out", () => {
    expect(resolvePermissionStatus(false, false, [])).toBe(PERMISSION_STATUS.ready);
  });
});

describe("resolveActionPermissions", () => {
  it("allows nothing while permissions are loading", () => {
    const granted: GrantedPermissions = { status: PERMISSION_STATUS.loading, can: () => true };
    expect(resolveActionPermissions("experiences", granted)).toEqual(allowedExcept(ALL, PERMISSION_STATUS.loading));
  });

  it("allows nothing when permissions are unavailable", () => {
    const granted: GrantedPermissions = { status: PERMISSION_STATUS.unavailable, can: () => true };
    expect(resolveActionPermissions("experiences", granted)).toEqual(allowedExcept(ALL, PERMISSION_STATUS.unavailable));
  });

  it("maps each action to its scope", () => {
    expect(resolveActionPermissions("experiences", ready("experiences:update", "experiences:delete"))).toEqual(
      allowedExcept([CRUD_ACTIONS.create]),
    );
  });

  it("enables exactly the granted resources for a mixed token", () => {
    const granted = ready(
      "education-levels:create", "education-levels:update", "education-levels:delete",
      "genders:create", "genders:update", "genders:delete",
    );
    expect(resolveActionPermissions("genders", granted)).toEqual(allowedExcept());
    expect(resolveActionPermissions("experiences", granted)).toEqual(allowedExcept(ALL));
  });
});

describe("disabledReason", () => {
  it("is undefined for an allowed action", () => {
    expect(disabledReason(allowedExcept(), true, CRUD_ACTIONS.create, "Gender")).toBeUndefined();
  });

  it("distinguishes loading, unavailable and denied", () => {
    const reason = (status: PermissionStatus) =>
      disabledReason(allowedExcept([CRUD_ACTIONS.create], status), false, CRUD_ACTIONS.create, "Gender");
    expect(reason(PERMISSION_STATUS.loading)).toBe(PERMISSION_MESSAGES.checking);
    expect(reason(PERMISSION_STATUS.unavailable)).toBe(PERMISSION_MESSAGES.unavailable);
    expect(reason(PERMISSION_STATUS.ready)).toBe("You don't have permission to add gender items.");
  });

  it("uses the user-facing word for each action", () => {
    const none = allowedExcept(ALL);
    expect(disabledReason(none, false, CRUD_ACTIONS.update, "Gender")).toBe("You don't have permission to edit gender items.");
    expect(disabledReason(none, false, CRUD_ACTIONS.delete, "Gender")).toBe("You don't have permission to delete gender items.");
  });
});

describe("permissionDeniedMessage", () => {
  it("joins several actions into one sentence", () => {
    expect(permissionDeniedMessage([CRUD_ACTIONS.update, CRUD_ACTIONS.delete], "Experience")).toBe(
      "You don't have permission to edit or delete experience items.",
    );
  });
});

describe("permissionsHint", () => {
  it("says nothing while loading or when everything is allowed", () => {
    expect(permissionsHint(allowedExcept([CRUD_ACTIONS.create], PERMISSION_STATUS.loading), "Gender")).toBeNull();
    expect(permissionsHint(allowedExcept(), "Gender")).toBeNull();
  });

  it("lists the missing actions", () => {
    expect(permissionsHint(allowedExcept([CRUD_ACTIONS.create]), "Experience")).toBe(
      "You don't have permission to add experience items.",
    );
    expect(permissionsHint(allowedExcept(ALL), "Experience")).toBe(
      "You don't have permission to add, edit or delete experience items.",
    );
  });

  it("uses the same message as the tooltips when permissions couldn't be loaded", () => {
    expect(permissionsHint(allowedExcept([], PERMISSION_STATUS.unavailable), "Gender")).toBe(PERMISSION_MESSAGES.unavailable);
  });
});

describe("permissionGuards", () => {
  const setup = (denied: CrudAction[]) => {
    const ui = { setEditError: vi.fn(), setDeleteTarget: vi.fn(), flashToast: vi.fn() };
    return { ui, guards: permissionGuards<string>(allowedExcept(denied), "Gender", ui) };
  };

  it("lets allowed actions through without side effects", () => {
    const { ui, guards } = setup([]);
    expect(guards.canStart(CRUD_ACTIONS.create)).toBe(true);
    expect(guards.canSave()).toBe(true);
    expect(guards.canConfirmDelete()).toBe(true);
    guards.requestDelete("row");
    expect(ui.setDeleteTarget).toHaveBeenCalledWith("row");
    expect(ui.setEditError).not.toHaveBeenCalled();
    expect(ui.flashToast).not.toHaveBeenCalled();
  });

  it("refuses to start a denied action", () => {
    expect(setup([CRUD_ACTIONS.create]).guards.canStart(CRUD_ACTIONS.create)).toBe(false);
  });

  it("refuses to save without update permission and says why inline", () => {
    const { ui, guards } = setup([CRUD_ACTIONS.update]);
    expect(guards.canSave()).toBe(false);
    expect(ui.setEditError).toHaveBeenCalledWith("You don't have permission to edit gender items.");
  });

  it("doesn't open the delete dialog without delete permission", () => {
    const { ui, guards } = setup([CRUD_ACTIONS.delete]);
    guards.requestDelete("row");
    expect(ui.setDeleteTarget).not.toHaveBeenCalled();
  });

  it("closes the dialog and explains when deleting is no longer allowed", () => {
    const { ui, guards } = setup([CRUD_ACTIONS.delete]);
    expect(guards.canConfirmDelete()).toBe(false);
    expect(ui.setDeleteTarget).toHaveBeenCalledWith(null);
    expect(ui.flashToast).toHaveBeenCalledWith("You don't have permission to delete gender items.");
  });
});
