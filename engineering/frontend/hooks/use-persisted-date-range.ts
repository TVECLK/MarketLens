"use client";

import { useEffect, useState } from "react";
import { getDataRangeStart, getDataRangeEnd, toISODate } from "@/lib/date-range";

interface UsePersistedDateRangeResult {
  fromDate: string;
  toDate: string;
  setFromDate: (value: string) => void;
  setToDate: (value: string) => void;
  rangeInvalid: boolean;
  restored: boolean;
}

function isValidDateString(value: unknown): value is string {
  return typeof value === "string" && /^\d{4}-\d{2}-\d{2}$/.test(value);
}

/**
 * Persists a from/to date range to sessionStorage under `storageKey`,
 * restoring it on mount if it's still within the last 3 years up to today.
 */
export function usePersistedDateRange(
  storageKey: string,
): UsePersistedDateRangeResult {
  const defaultFrom = `${getDataRangeEnd().getFullYear()}-01-01`;
  const defaultTo = toISODate(getDataRangeEnd());

  const [fromDate, setFromDate] = useState(defaultFrom);
  const [toDate, setToDate] = useState(defaultTo);
  const [restored, setRestored] = useState(false);

  // Restoring persisted state from sessionStorage (a client-only API,
  // unavailable during SSR) has to happen post-mount so the first client
  // render still matches the server-rendered defaults — this is the
  // standard hydration-safe pattern, not state derived from props/state
  // that could be computed during render instead.
  useEffect(() => {
    try {
      const saved = sessionStorage.getItem(storageKey);
      if (saved) {
        const { from, to } = JSON.parse(saved);
        if (
          isValidDateString(from) &&
          isValidDateString(to) &&
          from <= to &&
          from >= toISODate(getDataRangeStart()) &&
          to <= toISODate(getDataRangeEnd())
        ) {
          // eslint-disable-next-line react-hooks/set-state-in-effect -- see note above
          setFromDate(from);
           
          setToDate(to);
        }
      }
    } catch {
      // ignore corrupt storage — defaults stay in place
    }
    setRestored(true);
  }, [storageKey]);

  useEffect(() => {
    if (!restored) return;
    try {
      sessionStorage.setItem(
        storageKey,
        JSON.stringify({ from: fromDate, to: toDate }),
      );
    } catch {
      // storage unavailable (private mode, quota) — persistence is best-effort
    }
  }, [restored, storageKey, fromDate, toDate]);

  return {
    fromDate,
    toDate,
    setFromDate,
    setToDate,
    rangeInvalid: fromDate > toDate,
    restored,
  };
}
