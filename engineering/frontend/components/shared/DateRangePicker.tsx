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
  /** Shown in place of the hint when the range can't be used. */
  error?: string;
}

export default function DateRangePicker({
  fromDate,
  toDate,
  onFromDateChange,
  onToDateChange,
  minDate,
  maxDate,
  scrolled,
  error,
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
      {error && (
        <p role="alert" className="text-xs text-red-500 font-bold mb-2">
          {error}
        </p>
      )}
      <div className="flex items-center justify-between gap-3 pl-12">
        <div className="flex items-center">
          <div className="flex flex-col items-center justify-center gap-0.75 pr-8">
            <p className="text-2xl font-semibold text-gray-700 leading-none">5</p>
            <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest">Sources</p>
          </div>
          <div className="flex flex-col items-center justify-center gap-0.75 border-l-2 border-indigo-300 px-8">
            <p className="text-2xl font-semibold  text-gray-700 leading-none">200k+</p>
            <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest">Total Jobs</p>
          </div>
          <div className="flex flex-col items-center justify-center gap-0.75 border-l-2 border-indigo-300 pl-8">
            <p className="text-2xl font-semibold text-gray-700 leading-none">21</p>
            <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest">Industries</p>
          </div>
        </div>
      <div className="flex flex-wrap items-center gap-3">
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
            className="bg-transparent text-xs font-bold text-zinc-700 outline-none [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:rounded [&::-webkit-calendar-picker-indicator]:p-0.5 [&::-webkit-calendar-picker-indicator]:hover:bg-zinc-300 [&::-webkit-calendar-picker-indicator]:active:bg-zinc-400 [&::-webkit-calendar-picker-indicator]:transition-colors"
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
            className="bg-transparent text-xs font-bold text-zinc-700 outline-none [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:rounded [&::-webkit-calendar-picker-indicator]:p-0.5 [&::-webkit-calendar-picker-indicator]:hover:bg-zinc-300 [&::-webkit-calendar-picker-indicator]:active:bg-zinc-400 [&::-webkit-calendar-picker-indicator]:transition-colors"
          />
        </div>
      </div>
      </div>
    </div>
  );
}
