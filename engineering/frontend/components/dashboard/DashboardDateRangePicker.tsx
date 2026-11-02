"use client";

import DateRangePicker from "@/components/shared/DateRangePicker";
import { useScrolled } from "@/hooks/use-scrolled";
import { getMaxSelectableDate, getMinSelectableDate } from "@/lib/date-range";
import { useDateRange } from "@/providers/date-range-provider";

/** The date range picker wired to the shared date range, for the dashboard pages. */
export default function DashboardDateRangePicker() {
  const { fromDate, toDate, setFromDate, setToDate } = useDateRange();
  const scrolled = useScrolled();

  return (
    <DateRangePicker
      fromDate={fromDate}
      toDate={toDate}
      onFromDateChange={setFromDate}
      onToDateChange={setToDate}
      minDate={getMinSelectableDate()}
      maxDate={getMaxSelectableDate()}
      scrolled={scrolled}
    />
  );
}
