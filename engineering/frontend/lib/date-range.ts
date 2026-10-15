export const toISODate = (date: Date) => date.toISOString().slice(0, 10);

// Computed on each call, deliberately NOT as module-level constants: this
// module is evaluated once per server process (Next.js caches ES modules),
// so a top-level `new Date()` would freeze "today" at whatever moment the
// server first loaded it — every user's request would then see that same
// stale date instead of their own current one. Call these at the point of
// use (inside a hook or component body) instead of caching the result.
export function getDataRangeEnd(): Date {
  return new Date();
}

export function getDataRangeStart(): Date {
  const end = getDataRangeEnd();
  return new Date(end.getFullYear() - 3, 0, 1);
}
