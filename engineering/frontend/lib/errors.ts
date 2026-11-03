export function getErrorMessage(err: unknown): string | undefined;
export function getErrorMessage(err: unknown, fallback: string): string;
export function getErrorMessage(err: unknown, fallback?: string) {
  return err instanceof Error ? err.message : fallback;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function backendError(details: unknown): string | undefined {
  if (isRecord(details) && typeof details.error === "string") return details.error;
  // GET routes forward the backend body as raw text, which is usually JSON.
  if (typeof details === "string") {
    try {
      return backendError(JSON.parse(details));
    } catch {
      return undefined;
    }
  }
  return undefined;
}

/**
 * Picks a user-facing message from an error body returned by our /api routes.
 *
 * When the backend fails, the routes reply with a generic
 * `{ error: "Backend returned an error", details: <backend body> }`; the
 * useful message (e.g. "missing required scope: genders:create") is the
 * backend body's own `error`. The backend's `details` (raw validation / SQL
 * errors) are deliberately not shown.
 */
export function getApiErrorMessage(body: unknown, fallback: string): string {
  if (!isRecord(body)) return fallback;
  const fromBackend = backendError(body.details);
  if (fromBackend) return fromBackend;
  if (typeof body.error === "string" && body.error) return body.error;
  return fallback;
}
