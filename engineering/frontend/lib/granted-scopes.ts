/** A scope value as stored by Thunder or the SDK: a space-separated string or a list. */
export function parseScopes(value: unknown): string[] {
  if (Array.isArray(value)) return value.filter((scope): scope is string => typeof scope === "string");
  if (typeof value === "string") return value.split(/\s+/).filter(Boolean);
  return [];
}

/**
 * The scopes in an access token's `scope` (or `scopes`) claim. Decodes without verifying: the token
 * comes from our own signed session cookie, and the backend verifies it on every request.
 */
export function scopesFromAccessToken(accessToken: string | undefined): string[] {
  const payload = accessToken?.split(".")[1];
  if (!payload) return [];
  try {
    const base64 = payload.replace(/-/g, "+").replace(/_/g, "/");
    const bytes = Uint8Array.from(atob(base64), (char) => char.charCodeAt(0));
    const claims = JSON.parse(new TextDecoder().decode(bytes)) as Record<string, unknown>;
    return parseScopes(claims.scope ?? claims.scopes);
  } catch {
    return [];
  }
}

/**
 * Whether `granted` covers `required`, matching the Thunder SDK's rule: a scope covers itself and
 * everything under it, so `kpis` covers `kpis:access`.
 */
export function hasScope(granted: readonly string[], required: string): boolean {
  return granted.some((scope) => scope === required || required.startsWith(`${scope}:`));
}
