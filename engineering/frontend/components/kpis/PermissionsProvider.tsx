"use client";

import { createContext, useMemo, type ReactNode } from "react";
import { useGrantedPermissions, useThunderID } from "@thunderid/nextjs";
import { resolvePermissionStatus, type GrantedPermissions } from "@/lib/action-permissions";

export const GrantedPermissionsContext = createContext<GrantedPermissions | null>(null);

/**
 * Loads the signed-in user's granted permissions once and shares them with every KPI table.
 * Lives in the KPI layout, which stays mounted while switching tables, so switching doesn't
 * re-fetch or flicker the buttons.
 */
export default function PermissionsProvider({ children }: { children: ReactNode }) {
  const { can, isLoading, permissions } = useGrantedPermissions();
  // Computed by the server-side ThunderIDProvider from the session, so it is already correct on
  // the first client render (no brief "signed out" state that would flash a "no permission" hint).
  const { isSignedIn } = useThunderID();

  const value = useMemo<GrantedPermissions>(
    () => ({ status: resolvePermissionStatus(!!isSignedIn, isLoading, permissions), can }),
    [can, isLoading, isSignedIn, permissions],
  );

  return <GrantedPermissionsContext.Provider value={value}>{children}</GrantedPermissionsContext.Provider>;
}
