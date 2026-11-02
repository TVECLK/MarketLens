// sessionStorage key under which the selected date range is remembered.
export const DATE_RANGE_STORAGE_KEY = "lmis-date-range:v2";

export const toISODate = (date: Date) => {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
};

// This module is evaluated once per server process,
// so a top-level `new Date()` would freeze "today" at server start.
export function getMaxSelectableDate(): Date {
  return new Date();
}

export function getMinSelectableDate(): Date {
  const end = getMaxSelectableDate();
  return new Date(end.getFullYear() - 3, 0, 1);
}
