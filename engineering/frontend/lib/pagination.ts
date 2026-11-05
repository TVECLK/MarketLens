/**
 * Keeps a 1-based `page` within the pages that exist. `total` is the server's
 * row count, or undefined while it hasn't loaded yet (the page is left alone).
 * An empty list still has page 1.
 */
export function clampPage(page: number, total: number | undefined, pageSize: number): number {
  if (total === undefined) return page;
  const lastPage = Math.max(1, Math.ceil(total / pageSize));
  return Math.min(Math.max(1, page), lastPage);
}
