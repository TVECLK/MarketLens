import { getApiErrorMessage } from "@/lib/errors";

interface FetchJsonOptions {
  signal?: AbortSignal;
  /** Used when the error response carries no message of its own. */
  errorMessage?: string;
  /** The HTTP method. Defaults to GET. */
  method?: string;
  /** Sent as JSON, with a `Content-Type: application/json` header. */
  body?: unknown;
  /** What sends the request. Defaults to `fetch`; pass `apiFetch` to go through the mock switch. */
  fetcher?: typeof fetch;
}

/**
 * Requests `url` (GET unless `method` says otherwise) and returns the parsed JSON. A non-OK response
 * throws an Error with the message our /api routes put in the body, or `errorMessage` when the body
 * has none (or isn't JSON).
 */
export async function fetchJson<T>(url: string, options: FetchJsonOptions = {}): Promise<T> {
  const { signal, errorMessage = `Request failed: ${url}`, method, body, fetcher = fetch } = options;

  // Only the options that were given, so a plain GET is sent exactly as `fetch(url, { signal })`.
  const init: RequestInit = { signal };
  if (method) init.method = method;
  if (body !== undefined) {
    init.headers = { "Content-Type": "application/json" };
    init.body = JSON.stringify(body);
  }

  const res = await fetcher(url, init);
  if (!res.ok) {
    const errorBody = await res.json().catch(() => null);
    throw new Error(getApiErrorMessage(errorBody, errorMessage));
  }
  return res.json();
}
