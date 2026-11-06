import { getApiErrorMessage } from "@/lib/errors";

interface FetchJsonOptions {
  signal?: AbortSignal;
  /** Used when the error response carries no message of its own. */
  errorMessage?: string;
}

/**
 * GETs `url` and returns the parsed JSON. A non-OK response throws an Error
 * with the message our /api routes put in the body, or `errorMessage` when
 * the body has none (or isn't JSON).
 */
export async function fetchJson<T>(url: string, options: FetchJsonOptions = {}): Promise<T> {
  const { signal, errorMessage = `Request failed: ${url}` } = options;
  const res = await fetch(url, { signal });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(getApiErrorMessage(body, errorMessage));
  }
  return res.json();
}
