"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { getDataRangeEnd, toISODate } from "@/lib/date-range";

interface UseUrlAwareDateRangeResult {
  fromDate: string;
  toDate: string;
  setFromDate: (value: string) => void;
  setToDate: (value: string) => void;
  rangeInvalid: boolean;
}

/**
 * Same persisted date-range pattern as `usePersistedDateRange`, but for pages
 * reachable via a deep link carrying `?from=...&to=...` (e.g. the dashboard's
 * "See More" links): a URL date range takes priority and skips the
 * sessionStorage restore entirely. Restored sessionStorage values are used
 * as-is here (no format/bounds validation), unlike `usePersistedDateRange` —
 * that's the pre-existing behavior of this pattern, not an oversight here.
 */
export function useUrlAwareDateRange(storageKey: string): UseUrlAwareDateRangeResult {
  const searchParams = useSearchParams();
  const urlFrom = searchParams.get("from");
  const urlTo = searchParams.get("to");

  const defaultFrom = `${getDataRangeEnd().getFullYear()}-01-01`;
  const defaultTo = toISODate(getDataRangeEnd());

  const [fromDate, setFromDate] = useState(urlFrom ?? defaultFrom);
  const [toDate, setToDate] = useState(urlTo ?? defaultTo);
  const [restored, setRestored] = useState(false);

  // Restoring persisted state from sessionStorage (a client-only API,
  // unavailable during SSR) has to happen post-mount so the first client
  // render still matches the server-rendered defaults — this is the
  // standard hydration-safe pattern, not state derived from props/state
  // that could be computed during render instead.
  useEffect(() => {
    if (!urlFrom && !urlTo) {
      try {
        const saved = sessionStorage.getItem(storageKey);
        if (saved) {
          const { from, to } = JSON.parse(saved);
          // eslint-disable-next-line react-hooks/set-state-in-effect -- see note above
          if (from) setFromDate(from);
           
          if (to) setToDate(to);
        }
      } catch {
        // ignore corrupt storage — defaults/URL values stay in place
      }
    }
    setRestored(true);
    // Intentionally run once on mount only: a later change to the URL's
    // from/to params shouldn't retroactively override user interaction.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!restored) return;
    try {
      sessionStorage.setItem(storageKey, JSON.stringify({ from: fromDate, to: toDate }));
    } catch {
      // storage unavailable — persistence is best-effort
    }
  }, [restored, storageKey, fromDate, toDate]);

  return { fromDate, toDate, setFromDate, setToDate, rangeInvalid: fromDate > toDate };
}
