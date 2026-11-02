"use client";

import { createContext, useContext, useMemo, type ReactNode } from "react";
import { useUrlDateRange } from "@/hooks/use-url-date-range";

type DateRangeContextValue = ReturnType<typeof useUrlDateRange>;

const DateRangeContext = createContext<DateRangeContextValue | null>(null);

/**
 * Holds the one shared date range (backed by the URL) for every page below it,
 * so the picker and the pages read the same state instead of each running its own.
 */
export function DateRangeProvider({ children }: { children: ReactNode }) {
  const { fromDate, toDate, setFromDate, setToDate } = useUrlDateRange();
  const value = useMemo(
    () => ({ fromDate, toDate, setFromDate, setToDate }),
    [fromDate, toDate, setFromDate, setToDate],
  );

  return <DateRangeContext.Provider value={value}>{children}</DateRangeContext.Provider>;
}

export function useDateRange(): DateRangeContextValue {
  const context = useContext(DateRangeContext);
  if (!context) throw new Error("useDateRange must be used within a DateRangeProvider");
  return context;
}
