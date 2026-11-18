"use client";

import { clsx } from "clsx";
import { toISODate } from "@/lib/date-range";
import InfoTooltip from "@/components/shared/InfoTooltip";

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
        <div className="flex items-center pl-6">
          <div className="flex flex-col items-center justify-center gap-0.5 pr-8">
            <p className="text-2xl font-semibold text-gray-700 leading-none">5</p>
            <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest flex items-center gap-1">
              Sources <InfoTooltip text="Number of job listing sources scraped" className="text-zinc-400 hover:text-zinc-600" />
            </p>
          </div>
          <div className="flex flex-col items-center justify-center gap-0.5 border-l-2 border-indigo-300 px-8">
            <p className="text-2xl font-semibold text-gray-700 leading-none">200k+</p>
            <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest flex items-center gap-1">
              Total Jobs <InfoTooltip text="Total job vacancies collected across all sources" className="text-zinc-400 hover:text-zinc-600" />
            </p>
          </div>
          <div className="flex flex-col items-center justify-center gap-0.5 border-l-2 border-indigo-300 pl-8">
            <p className="text-2xl font-semibold text-gray-700 leading-none">21</p>
            <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest flex items-center gap-1">
              Industries <InfoTooltip text="Number of industries classified under SLSIC" className="text-zinc-400 hover:text-zinc-600" />
            </p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2 bg-white px-3 py-2 rounded-md">
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

        <div className="flex items-center gap-2 bg-white px-3 py-2 rounded-md">
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
