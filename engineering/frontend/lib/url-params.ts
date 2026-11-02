// Single place that defines how page state is encoded in the URL, so every
// page produces and understands the same shareable links:
//
//   ?from=YYYY-MM-DD&to=YYYY-MM-DD          date range (all pages)
//   &<level-key>=<id>                       one param per hierarchy level
//                                           (e.g. major-group=2&sub-major-group=11)
//   &view=chart|table                       analysis view mode
//   &tab=analytics|skills                   occupations page tab

export const URL_PARAMS = {
  from: "from",
  to: "to",
  view: "view",
  tab: "tab",
} as const;

interface ParamReader {
  get(name: string): string | null;
}

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

// YYYY-MM-DD that is also a real calendar day (rejects 2026-02-30). Built from
// numeric parts with Date.UTC: `new Date("2026-02-30")` doesn't reject it, it
// silently rolls over to March 2 in V8.
function isCalendarDate(value: string): boolean {
  if (!DATE_PATTERN.test(value)) return false;
  const [year, month, day] = value.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return (
    date.getUTCFullYear() === year &&
    date.getUTCMonth() === month - 1 &&
    date.getUTCDate() === day
  );
}

export function isDateString(value: unknown): value is string {
  return typeof value === "string" && isCalendarDate(value.trim());
}

export function isUsableDateParam(value: string | null): value is string {
  if (value === null) return false;
  const trimmed = value.trim();
  return trimmed === "" || isCalendarDate(trimmed);
}

/** Returns the param's value if it is one of `allowed`, otherwise `fallback`. */
export function parseEnumParam<T extends string>(
  value: string | null,
  allowed: readonly T[],
  fallback: T,
): T {
  return allowed.find((candidate) => candidate === value) ?? fallback;
}

/**
 * Reads the selected id at each hierarchy level, in order. Stops at the first
 * level whose param is missing or not a non-negative integer — a deeper level
 * without its parent can't be resolved, so it is ignored.
 */
export function parseHierarchyIds(params: ParamReader, levelKeys: readonly string[]): number[] {
  const ids: number[] = [];
  for (const key of levelKeys) {
    const raw = params.get(key);
    const id = raw !== null && /^\d+$/.test(raw) ? Number(raw) : NaN;
    if (!Number.isSafeInteger(id)) break;
    ids.push(id);
  }
  return ids;
}

/** Inverse of `parseHierarchyIds`: sets selected levels and clears deeper ones. */
export function serializeHierarchyIds(
  ids: readonly number[],
  levelKeys: readonly string[],
): Record<string, string | null> {
  const updates: Record<string, string | null> = {};
  levelKeys.forEach((key, index) => {
    updates[key] = index < ids.length ? String(ids[index]) : null;
  });
  return updates;
}
