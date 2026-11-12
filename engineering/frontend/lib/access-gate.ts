import { hasScope } from "@/lib/granted-scopes";
import { PAGE_SCOPES } from "@/lib/scopes";

/**
 * Where Thunder sends the browser after the login form (`NEXT_PUBLIC_THUNDERID_AFTER_SIGN_IN_URL`).
 * Shows a loading screen while the login completes, then the proxy routes the user on.
 */
export const AUTH_PATH = "/auth";
/** Where a signed-in user without app access is sent to be signed out. */
export const ACCESS_DENIED_PATH = "/auth/access-denied";

/**
 * The admin pages and the page scope each needs, in the order a signed-in user is sent to them:
 * the first one they can open is their admin home.
 */
const ADMIN_PAGES: readonly { path: string; scope: string }[] = [
  { path: "/kpis", scope: PAGE_SCOPES.kpis },
  { path: "/data-sources", scope: PAGE_SCOPES.datasource },
];

/** Whether the user can use the app at all: they hold at least one page scope. */
export function hasAppAccess(grantedScopes: readonly string[]): boolean {
  return Object.values(PAGE_SCOPES).some((scope) => hasScope(grantedScopes, scope));
}

/** The page scope `pathname` needs (its own path or anything under it), if it's an admin page. */
export function requiredPageScope(pathname: string): string | undefined {
  return ADMIN_PAGES.find(({ path }) => pathname === path || pathname.startsWith(`${path}/`))?.scope;
}

/** The first admin page the user can open, or undefined when they have no app access. */
export function adminHomePath(grantedScopes: readonly string[]): string | undefined {
  return ADMIN_PAGES.find(({ scope }) => hasScope(grantedScopes, scope))?.path;
}

/**
 * What the proxy does with a request:
 * - `allow`: let it through;
 * - `sign-in`: a signed-out visitor opened a protected page;
 * - `deny`: a signed-in user with no page scope at all. Signing in to this app is only for users who
 *   can use it, so such a session is ended (both here and in Thunder) instead of kept;
 * - `to-admin`: login finished and the user has access, or the user opened an admin page they don't
 *   have the scope for; either way they go to `adminHomePath`, the first admin page they can open;
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
  const hasAccess = request.isSignedIn && hasAppAccess(request.grantedScopes);

  if (request.pathname === AUTH_PATH) {
    // Not signed in yet: show the loading page while the SDK completes the login from the `code`.
    if (!request.isSignedIn) return request.hasAuthCode ? "allow" : "to-home";
    return hasAccess ? "to-admin" : "deny";
  }

  if (!request.isSignedIn) return request.isProtectedPath ? "sign-in" : "allow";
  // The sign-out page itself (and its sign-out server action) must stay reachable.
  if (request.pathname === ACCESS_DENIED_PATH) return "allow";
  if (!hasAccess) return "deny";
  // Has app access, so `adminHomePath` names a page they can open: this redirect can't loop.
  const pageScope = requiredPageScope(request.pathname);
  if (pageScope && !hasScope(request.grantedScopes, pageScope)) return "to-admin";
  return "allow";
}
