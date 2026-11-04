import type { SectionError } from "@/types/bff";

/** How long each section's backend call may take before it is aborted. */
const SECTION_TIMEOUT_MS = 3000;

/** The status an optional section is expected to fail with (e.g. `children` on a leaf level). */
const EXPECTED_OPTIONAL_STATUS = 400;

export interface SectionRequest {
  key: string;
  url: string;
  optional?: boolean;
}

export interface AggregateResult {
  /** Parsed JSON per key, or `null` when that section failed. */
  data: Record<string, unknown>;
  /** One entry per failed section, except an optional section's expected 400. */
  errors: Record<string, SectionError>;
  /** True when every required section failed. */
  allFailed: boolean;
  /** Status and body of the first failed required section, to forward when everything failed. */
  firstError: { status: number; body: unknown } | null;
}

/**
 * Calls every section's backend URL in parallel and reports each outcome
 * separately, so one failing endpoint doesn't take the others down with it.
 * Never throws.
 */
export async function fetchSections(
  requests: SectionRequest[],
  fetchFn: typeof fetch = fetch,
): Promise<AggregateResult> {
  const settled = await Promise.allSettled(requests.map(({ url }) => fetchFn(url, { signal: AbortSignal.timeout(SECTION_TIMEOUT_MS) })));

  const data: Record<string, unknown> = {};
  const errors: Record<string, SectionError> = {};
  let firstError: AggregateResult["firstError"] = null;

  const fail = (key: string, optional: boolean | undefined, status: number, body: unknown, message: string) => {
    data[key] = null;
    if (optional && status === EXPECTED_OPTIONAL_STATUS) return;
    console.error(`[bff] ${key} failed (${status}): ${message}`);
    errors[key] = { status, message };
    if (!optional) firstError ??= { status, body };
  };

  for (let i = 0; i < requests.length; i++) {
    const { key, optional } = requests[i];
    const result = settled[i];

    if (result.status === "rejected") {
      const details = result.reason instanceof Error ? result.reason.message : String(result.reason);
      if (result.reason instanceof Error && result.reason.name === "TimeoutError") {
        fail(key, optional, 504, { error: `Timed out waiting for the backend for ${key}`, details }, details);
      } else {
        fail(key, optional, 502, { error: `Failed to reach the backend for ${key}`, details }, details);
      }
      continue;
    }

    const res = result.value;
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      const message = body?.error ?? body?.details ?? `Backend returned an error from ${key}`;
      fail(key, optional, res.status, body ?? { error: message }, message);
      continue;
    }

    try {
      data[key] = await res.json();
    } catch {
      const message = `Invalid response from ${key}`;
      fail(key, optional, 502, { error: message }, message);
    }
  }

  const requiredKeys = requests.filter((r) => !r.optional).map((r) => r.key);
  const requiredFailedCount = requiredKeys.filter((key) => key in errors).length;
  const allFailed = requiredKeys.length > 0 && requiredFailedCount === requiredKeys.length;

  return { data, errors, allFailed, firstError };
}
