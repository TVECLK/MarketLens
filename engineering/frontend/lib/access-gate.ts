import { hasScope } from "@/lib/granted-scopes";
import { PAGE_SCOPES } from "@/lib/scopes";

/**
 * Where Thunder sends the browser after the login form (`NEXT_PUBLIC_THUNDERID_AFTER_SIGN_IN_URL`).
 * Shows a loading screen while the login completes, then the proxy routes the user on.
 */
export const AUTH_PATH = "/auth";
/** Where a signed-in user without app access is sent to be signed out. */
export const ACCESS_DENIED_PATH = "/auth/access-denied";
/** Where a user with app access lands after signing in. */
export const ADMIN_HOME_PATH = "/kpis";

/**
 * What the proxy does with a request:
 * - `allow`: let it through;
 * - `sign-in`: a signed-out visitor opened a protected page;
 * - `deny`: a signed-in user without `kpis:access`. Signing in to this app is only for users who
 *   can use it, so such a session is ended (both here and in Thunder) instead of kept;
 * - `to-admin`: login finished and the user has access;
 * - `to-home`: `/auth` opened with no login in progress.
 */
export type GateDecision = "allow" | "sign-in" | "deny" | "to-admin" | "to-home";

export function gateDecision(request: {
  isSignedIn: boolean;
  grantedScopes: readonly string[];
  pathname: string;
  isProtectedPath: boolean;
  /** The URL carries an OAuth `code`, i.e. Thunder just sent the user back. */
  hasAuthCode: boolean;
}): GateDecision {
  const hasAccess = request.isSignedIn && hasScope(request.grantedScopes, PAGE_SCOPES.kpis);

  if (request.pathname === AUTH_PATH) {
    // Not signed in yet: show the loading page while the SDK completes the login from the `code`.
    if (!request.isSignedIn) return request.hasAuthCode ? "allow" : "to-home";
    return hasAccess ? "to-admin" : "deny";
  }

  if (!request.isSignedIn) return request.isProtectedPath ? "sign-in" : "allow";
  // The sign-out page itself (and its sign-out server action) must stay reachable.
  if (request.pathname === ACCESS_DENIED_PATH) return "allow";
  return hasAccess ? "allow" : "deny";
}
