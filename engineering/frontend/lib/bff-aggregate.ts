import type { SectionError } from "@/types/bff";

export interface SectionRequest {
  /** Key the section's data is returned under. */
  key: string;
  url: string;
  /** A failure is expected and silent (e.g. `children` on a leaf level): no error is recorded. */
  optional?: boolean;
}

export interface AggregateResult {
  /** Parsed JSON per key, or `null` when that section failed. */
  data: Record<string, unknown>;
  /** One entry per failed required section. */
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
  const settled = await Promise.allSettled(requests.map(({ url }) => fetchFn(url)));

  const data: Record<string, unknown> = {};
  const errors: Record<string, SectionError> = {};
  let firstError: AggregateResult["firstError"] = null;

  const fail = (key: string, optional: boolean | undefined, status: number, body: unknown, message: string) => {
    data[key] = null;
    if (optional) return;
    console.error(`[bff] ${key} failed (${status}): ${message}`);
    errors[key] = { status, message };
    firstError ??= { status, body };
  };

  for (let i = 0; i < requests.length; i++) {
    const { key, optional } = requests[i];
    const result = settled[i];

    if (result.status === "rejected") {
      const details = result.reason instanceof Error ? result.reason.message : String(result.reason);
      fail(key, optional, 502, { error: `Failed to reach the backend for ${key}`, details }, details);
      continue;
    }

    const res = result.value;
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      const message = body?.error ?? body?.details ?? `Go backend returned an error from ${key}`;
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

  const requiredCount = requests.filter((r) => !r.optional).length;
  const allFailed = requiredCount > 0 && Object.keys(errors).length === requiredCount;

  return { data, errors, allFailed, firstError };
}
