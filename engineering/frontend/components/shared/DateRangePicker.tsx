"use client";

import { clsx } from "clsx";
import { toISODate } from "@/lib/date-range";

interface DateRangePickerProps {
  fromDate: string;
  toDate: string;
  onFromDateChange: (value: string) => void;
  onToDateChange: (value: string) => void;
  minDate: Date;
  maxDate: Date;
  scrolled: boolean;
}

export default function DateRangePicker({
  fromDate,
  toDate,
  onFromDateChange,
  onToDateChange,
  minDate,
  maxDate,
  scrolled,
}: DateRangePickerProps) {
  return (
    <div
      className={clsx(
        "sticky top-0 z-30 px-4 md:px-8 py-3 md:py-4 transition-all duration-300",
        "bg-white/60 backdrop-blur-md backdrop-saturate-150",
        scrolled
          ? "shadow-[0_4px_20px_-4px_rgba(0,0,0,0.08)] border-b border-zinc-200/60"
          : "border-b border-transparent",
      )}
    >
      <p className="text-xs text-zinc-400 font-medium">Select a date range</p>
      <div className="flex flex-wrap items-center gap-3 mt-3">
        <div className="flex items-center gap-2 bg-zinc-100 px-3 py-2 rounded-md">
          <label className="text-[10px] font-black uppercase tracking-wider text-zinc-400">
            From
          </label>
          <input
            type="date"
            value={fromDate}
            min={toISODate(minDate)}
            max={toDate}
            onChange={(e) => onFromDateChange(e.target.value)}
            className="bg-transparent text-xs font-bold text-zinc-700 outline-none [&::-webkit-calendar-picker-indicator]:cursor-pointer"
          />
        </div>

        <span className="text-zinc-400 font-black text-sm select-none" aria-hidden>
          →
        </span>

        <div className="flex items-center gap-2 bg-zinc-100 px-3 py-2 rounded-md">
          <label className="text-[10px] font-black uppercase tracking-wider text-zinc-400">
            To
          </label>
          <input
            type="date"
            value={toDate}
            min={fromDate}
            max={toISODate(maxDate)}
            onChange={(e) => onToDateChange(e.target.value)}
            className="bg-transparent text-xs font-bold text-zinc-700 outline-none [&::-webkit-calendar-picker-indicator]:cursor-pointer"
          />
        </div>
      </div>
    </div>
  );
}
