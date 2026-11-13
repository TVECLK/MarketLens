/**
 * Helpers for source names. An auto source's name must exactly match its crawler's SOURCE_NAME
 * (case-sensitive), so the add form shows the name exactly as it will be saved and warns about
 * the mistakes that break that match.
 */

/** The name as the backend saves it: leading and trailing whitespace removed. */
export function normalizeSourceName(raw: string): string {
  return raw.trim();
}

/** Whether saving will strip whitespace from the start or end of what was typed. */
export function hasOuterWhitespace(raw: string): boolean {
  const saved = normalizeSourceName(raw);
  return saved !== "" && saved !== raw;
}

/** Whether the saved name contains whitespace, which rarely belongs in a crawler's SOURCE_NAME. */
export function hasInnerWhitespace(raw: string): boolean {
  return /\s/.test(normalizeSourceName(raw));
}

/** The saved name with each whitespace character shown as `·`, so it can't go unnoticed. */
export function visibleWhitespace(raw: string): string {
  return normalizeSourceName(raw).replace(/\s/g, "·");
}

/** An existing source with exactly this (saved) name, if any. */
export function findExactNameMatch<T extends { name: string }>(raw: string, sources: readonly T[]): T | undefined {
  const name = normalizeSourceName(raw);
  return sources.find((source) => source.name === name);
}

/**
 * An existing source whose name differs from this one only in letter case, e.g. "ikman" vs
 * "Ikman". Names are case-sensitive, so that would be a second source, usually by mistake.
 */
export function findCaseOnlyNameMatch<T extends { name: string }>(raw: string, sources: readonly T[]): T | undefined {
  const name = normalizeSourceName(raw);
  if (!name) return undefined;
  const lower = name.toLowerCase();
  return sources.find((source) => source.name !== name && source.name.toLowerCase() === lower);
}

/** Whether what the admin typed confirms deleting `name`: an exact, case-sensitive match. */
export function confirmsSourceName(typed: string, name: string): boolean {
  return normalizeSourceName(typed) === name;
}
