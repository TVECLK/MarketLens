"use client";

import { toISODate } from "@/lib/date-range";

interface DateRangePickerProps {
  fromDate: string;
  toDate: string;
  onFromDateChange: (value: string) => void;
  onToDateChange: (value: string) => void;
  minDate: Date;
  maxDate: Date;
}

/**
 * Mobile: the field fills its half of the row (`w-full min-w-0`, also needed for iOS Safari, which
 * otherwise gives date inputs a fixed minimum width) and uses 16px text, because iOS zooms the page
 * in when a field under 16px is focused. From `md`: the compact desktop look.
 */
const DATE_INPUT_CLASS =
  "w-full min-w-0 md:w-auto bg-transparent text-base md:text-xs font-bold text-zinc-700 outline-none [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:rounded [&::-webkit-calendar-picker-indicator]:p-0.5 [&::-webkit-calendar-picker-indicator]:hover:bg-zinc-300 [&::-webkit-calendar-picker-indicator]:active:bg-zinc-400 [&::-webkit-calendar-picker-indicator]:transition-colors";

/** Label above the date on mobile, beside it from `md`. */
const FIELD_CLASS =
  "flex flex-1 md:flex-none min-w-0 flex-col md:flex-row items-start md:items-center gap-0.5 md:gap-2 bg-white px-3 py-1.5 md:py-2 rounded-md";

const LABEL_CLASS = "text-[10px] font-black uppercase tracking-wider text-zinc-400";

/** The "From → To" date inputs. Layout, sticky bar and error message are up to the caller. */
export default function DateRangePicker({
  fromDate,
  toDate,
  onFromDateChange,
  onToDateChange,
  minDate,
  maxDate,
}: DateRangePickerProps) {
  return (
    // Mobile: one full-width row, the two fields sharing it equally. From `md`: sized to content.
    <div className="flex w-full md:w-auto items-center gap-2 md:gap-3">
      <label className={FIELD_CLASS}>
        <span className={LABEL_CLASS}>From</span>
        <input
          type="date"
          value={fromDate}
          min={toISODate(minDate)}
          max={toDate}
          onChange={(e) => onFromDateChange(e.target.value)}
          className={DATE_INPUT_CLASS}
        />
      </label>

      <span className="shrink-0 text-zinc-400 font-black text-sm select-none" aria-hidden>
        →
      </span>

      <label className={FIELD_CLASS}>
        <span className={LABEL_CLASS}>To</span>
        <input
          type="date"
          value={toDate}
          min={fromDate}
          max={toISODate(maxDate)}
          onChange={(e) => onToDateChange(e.target.value)}
          className={DATE_INPUT_CLASS}
        />
      </label>
    </div>
  );
}
