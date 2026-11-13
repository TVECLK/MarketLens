/**
 * Helpers for source names. Names are saved in lowercase ("Ikman" → "ikman") and shown with a
 * capital first letter ("ikman" → "Ikman"). The add form shows the name exactly as it will be saved
 * and warns about mistakes that would break an auto source's match with its crawler.
 */

/** The name as it is saved: leading and trailing whitespace removed, all lowercase. */
export function normalizeSourceName(raw: string): string {
  return raw.trim().toLowerCase();
}

/** A saved name as it is shown: first letter capitalized, e.g. "ikman" → "Ikman". */
export function displaySourceName(name: string): string {
  return name.charAt(0).toUpperCase() + name.slice(1);
}

/** Whether saving will strip whitespace from the start or end of what was typed. */
export function hasOuterWhitespace(raw: string): boolean {
  const trimmed = raw.trim();
  return trimmed !== "" && trimmed !== raw;
}

/** Whether saving will lowercase any letters that were typed. */
export function hasUppercase(raw: string): boolean {
  const trimmed = raw.trim();
  return trimmed !== trimmed.toLowerCase();
}

/** Whether the saved name contains whitespace, which rarely belongs in a crawler's SOURCE_NAME. */
export function hasInnerWhitespace(raw: string): boolean {
  return /\s/.test(normalizeSourceName(raw));
}

/** The saved name with each whitespace character shown as `·`, so it can't go unnoticed. */
export function visibleWhitespace(raw: string): string {
  return normalizeSourceName(raw).replace(/\s/g, "·");
}

/**
 * An existing source that would have the same saved name, if any. Compared in lowercase, so names
 * saved before lowercasing (e.g. "Ikman") still count as taken.
 */
export function findExactNameMatch<T extends { name: string }>(raw: string, sources: readonly T[]): T | undefined {
  const name = normalizeSourceName(raw);
  return sources.find((source) => normalizeSourceName(source.name) === name);
}

/** Whether what the admin typed confirms deleting `name`. Letter case is ignored, as when saving. */
export function confirmsSourceName(typed: string, name: string): boolean {
  return normalizeSourceName(typed) === normalizeSourceName(name);
}
