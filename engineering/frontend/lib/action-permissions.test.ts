import { describe, expect, it } from "vitest";
import {
  disabledReason,
  PERMISSION_MESSAGES,
  PERMISSION_STATUS,
  permissionDeniedMessage,
  permissionsHint,
  resolveActionPermissions,
  resolvePermissionStatus,
  type ActionPermissions,
  type GrantedPermissions,
} from "@/lib/action-permissions";
import { CRUD_ACTIONS } from "@/lib/scopes";

const ready = (...granted: string[]): GrantedPermissions => ({
  status: PERMISSION_STATUS.ready,
  can: (permission) => granted.includes(permission),
});

const allowed = (overrides: Partial<ActionPermissions> = {}): ActionPermissions => ({
  status: PERMISSION_STATUS.ready,
  canCreate: true,
  canUpdate: true,
  canDelete: true,
  ...overrides,
});

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
    expect(resolveActionPermissions("experiences", granted)).toEqual({
      status: PERMISSION_STATUS.loading,
      canCreate: false,
      canUpdate: false,
      canDelete: false,
    });
  });

  it("allows nothing when permissions are unavailable", () => {
    const granted: GrantedPermissions = { status: PERMISSION_STATUS.unavailable, can: () => true };
    expect(resolveActionPermissions("experiences", granted).canCreate).toBe(false);
  });

  it("maps each action to its scope", () => {
    expect(resolveActionPermissions("experiences", ready("experiences:update", "experiences:delete"))).toEqual(
      allowed({ canCreate: false }),
    );
  });

  it("enables exactly the granted resources for a mixed token", () => {
    const granted = ready(
      "education-levels:create", "education-levels:update", "education-levels:delete",
      "genders:create", "genders:update", "genders:delete",
    );
    expect(resolveActionPermissions("genders", granted)).toEqual(allowed());
    expect(resolveActionPermissions("experiences", granted)).toEqual(
      allowed({ canCreate: false, canUpdate: false, canDelete: false }),
    );
  });
});

describe("disabledReason", () => {
  it("is undefined for an allowed action", () => {
    expect(disabledReason(allowed(), true, CRUD_ACTIONS.create, "Gender")).toBeUndefined();
  });

  it("distinguishes loading, unavailable and denied", () => {
    const denied = { canCreate: false };
    const reason = (status: ActionPermissions["status"]) =>
      disabledReason(allowed({ ...denied, status }), false, CRUD_ACTIONS.create, "Gender");
    expect(reason(PERMISSION_STATUS.loading)).toBe(PERMISSION_MESSAGES.checking);
    expect(reason(PERMISSION_STATUS.unavailable)).toBe(PERMISSION_MESSAGES.unavailable);
    expect(reason(PERMISSION_STATUS.ready)).toBe("You don't have permission to add gender items.");
  });

  it("uses the user-facing word for each action", () => {
    const none = allowed({ canCreate: false, canUpdate: false, canDelete: false });
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
    expect(permissionsHint(allowed({ status: PERMISSION_STATUS.loading, canCreate: false }), "Gender")).toBeNull();
    expect(permissionsHint(allowed(), "Gender")).toBeNull();
  });

  it("lists the missing actions", () => {
    expect(permissionsHint(allowed({ canCreate: false }), "Experience")).toBe(
      "You don't have permission to add experience items.",
    );
    expect(permissionsHint(allowed({ canCreate: false, canUpdate: false, canDelete: false }), "Experience")).toBe(
      "You don't have permission to add, edit or delete experience items.",
    );
  });

  it("uses the same message as the tooltips when permissions couldn't be loaded", () => {
    expect(permissionsHint(allowed({ status: PERMISSION_STATUS.unavailable }), "Gender")).toBe(PERMISSION_MESSAGES.unavailable);
  });
});
