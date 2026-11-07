"use client";

import { useContext } from "react";
import { GrantedPermissionsContext } from "@/components/kpis/PermissionsProvider";
import { resolveActionPermissions, type ActionPermissions } from "@/lib/action-permissions";
import type { ResourceScope } from "@/lib/scopes";

/**
 * The CRUD actions the signed-in user may perform on a resource, from the permissions Thunder
 * granted in their access token. Must be used under `PermissionsProvider`.
 */
export function useActionPermissions(scope: ResourceScope): ActionPermissions {
  const granted = useContext(GrantedPermissionsContext);
  if (!granted) {
    throw new Error("useActionPermissions must be used inside <PermissionsProvider>");
  }
  return resolveActionPermissions(scope, granted);
}
