"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  DATE_RANGE_STORAGE_KEY,
  getMinSelectableDate,
  getMaxSelectableDate,
  toISODate,
} from "@/lib/date-range";
import { URL_PARAMS, isDateString, isUsableDateParam } from "@/lib/url-params";
import { useUpdateSearchParams } from "@/hooks/use-update-search-params";

interface UseUrlDateRangeResult {
  fromDate: string;
  toDate: string;
  setFromDate: (value: string) => void;
  setToDate: (value: string) => void;
}

interface DateRange {
  from: string;
  to: string;
}

function getDefaultRange(): DateRange {
  return {
    from: `${getMaxSelectableDate().getFullYear()}-01-01`,
    to: toISODate(getMaxSelectableDate()),
  };
}

// A URL date the picker could have produced: a real date within the
// selectable bounds, or "" (a cleared input). Anything else (a hand-edited or
// stale link) is treated as missing and replaced below.
function isUsableUrlDate(value: string | null): value is string {
  if (!isUsableDateParam(value)) return false;
  const trimmed = value.trim();
  return (
    trimmed === "" ||
    (trimmed >= toISODate(getMinSelectableDate()) && trimmed <= toISODate(getMaxSelectableDate()))
  );
}

// The last range saved in this browser session, if it is still valid. Reads
// sessionStorage (client-only), so call it from effects and handlers only.
function readSavedRange(storageKey: string): DateRange | null {
  try {
    const saved = sessionStorage.getItem(storageKey);
    if (!saved) return null;
    const parsed = JSON.parse(saved);
    if (!isDateString(parsed.from) || !isDateString(parsed.to)) return null;
    const from = parsed.from.trim();
    const to = parsed.to.trim();
    const inBounds =
      from <= to &&
      from >= toISODate(getMinSelectableDate()) &&
      to <= toISODate(getMaxSelectableDate());
    return inBounds ? { from, to } : null;
  } catch {
    // ignore corrupt storage — defaults stay in place
    return null;
  }
}

/**
 * The date range lives in the URL (`?from=YYYY-MM-DD&to=YYYY-MM-DD`) so any
 * snapshot of a page can be shared as a link, and a range in the URL always
 * wins. When the URL carries no (valid) range, the last range used in this
 * browser session (sessionStorage) or the current-year default is used, and
 * is then written back to the URL so the address bar is always a complete,
 * shareable snapshot.
 */
export function useUrlDateRange(
  storageKey: string = DATE_RANGE_STORAGE_KEY,
): UseUrlDateRangeResult {
  const searchParams = useSearchParams();
  const updateSearchParams = useUpdateSearchParams();
  const rawFrom = searchParams.get(URL_PARAMS.from);
  const rawTo = searchParams.get(URL_PARAMS.to);

  const [fallback, setFallback] = useState(getDefaultRange);
  const [restored, setRestored] = useState(false);
  const initialRangeChecked = useRef(false);

  const fromDate = isUsableUrlDate(rawFrom) ? rawFrom.trim() : fallback.from;
  const toDate = isUsableUrlDate(rawTo) ? rawTo.trim() : fallback.to;

  // Restoring from sessionStorage (a client-only API, unavailable during SSR)
  // has to happen post-mount so the first client render matches the server's.
  useEffect(() => {
    const saved = readSavedRange(storageKey);
    // eslint-disable-next-line react-hooks/set-state-in-effect -- client-only storage, see note above
    if (saved) setFallback(saved);
    setRestored(true);
  }, [storageKey]);

  // Make the address bar reflect the effective range whenever the URL lacks
  // one — on load, or when navigating to a page without a range (e.g. a link
  // to "/"). The saved range is read fresh here, since it may have changed
  // since this hook mounted.
  useEffect(() => {
    if (!restored) return;
    const missingFrom = !isUsableUrlDate(rawFrom);
    const missingTo = !isUsableUrlDate(rawTo);
    if (!missingFrom && !missingTo) return;

    const base = readSavedRange(storageKey) ?? getDefaultRange();
    const updates: Record<string, string> = {};
    if (missingFrom) updates[URL_PARAMS.from] = base.from;
    if (missingTo) updates[URL_PARAMS.to] = base.to;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- client-only storage, see note above
    setFallback(base);
    updateSearchParams(updates);
  }, [restored, storageKey, rawFrom, rawTo, updateSearchParams]);

  // A range typed into the URL with from > to can't be corrected in the picker
  // (its min/max conflict), so reset it to the default. Checked once on load
  // only: later edits are made through the picker.
  useEffect(() => {
    if (!restored || initialRangeChecked.current) return;
    initialRangeChecked.current = true;
    if (fromDate > toDate) {
      updateSearchParams({ [URL_PARAMS.from]: fallback.from, [URL_PARAMS.to]: fallback.to });
    }
  }, [restored, fromDate, toDate, fallback, updateSearchParams]);

  useEffect(() => {
    if (!restored || !fromDate || !toDate || fromDate > toDate) return;
    try {
      sessionStorage.setItem(storageKey, JSON.stringify({ from: fromDate, to: toDate }));
    } catch {
      // storage unavailable — persistence is best-effort
    }
  }, [restored, storageKey, fromDate, toDate]);

  const setFromDate = useCallback(
    (value: string) => updateSearchParams({ [URL_PARAMS.from]: value }),
    [updateSearchParams],
  );
  const setToDate = useCallback(
    (value: string) => updateSearchParams({ [URL_PARAMS.to]: value }),
    [updateSearchParams],
  );

  return { fromDate, toDate, setFromDate, setToDate };
}
