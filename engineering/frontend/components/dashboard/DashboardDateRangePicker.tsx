"use client";

import { clsx } from "clsx";
import DateRangePicker from "@/components/shared/DateRangePicker";
import DashboardStatCards from "@/components/dashboard/DashboardStatCards";
import { useScrolled } from "@/hooks/use-scrolled";
import { getMaxSelectableDate, getMinSelectableDate, isValidDateRange } from "@/lib/date-range";
import { useDateRange } from "@/providers/date-range-provider";

/** The sticky bar at the top of the dashboard pages: stat cards and the shared date range picker. */
export default function DashboardDateRangePicker() {
  const { fromDate, toDate, setFromDate, setToDate } = useDateRange();
  const scrolled = useScrolled();
  const error = isValidDateRange(fromDate, toDate)
    ? undefined
    : "Select both a start and an end date, with the start on or before the end.";

  return (
    <div
      className={clsx(
        "sticky top-0 z-30 px-4 md:px-8 py-3 md:py-4 transition-all duration-300",
        "bg-indigo-50/60 backdrop-blur-md backdrop-saturate-150",
        scrolled
          ? "shadow-[0_4px_20px_-4px_rgba(0,0,0,0.08)] border-b border-zinc-200/60"
          : "border-b border-transparent",
      )}
    >
      {error && (
        <p role="alert" className="text-xs text-red-500 font-bold mb-2">
          {error}
        </p>
      )}
      <div className="flex flex-wrap md:flex-nowrap items-center justify-between gap-3">
        <DashboardStatCards />
        <DateRangePicker
          fromDate={fromDate}
          toDate={toDate}
          onFromDateChange={setFromDate}
          onToDateChange={setToDate}
          minDate={getMinSelectableDate()}
          maxDate={getMaxSelectableDate()}
        />
      </div>
    </div>
  );
}
