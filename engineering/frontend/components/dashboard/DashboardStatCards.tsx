"use client";

import InfoTooltip from "@/components/shared/InfoTooltip";
import { useDashboardCardData } from "@/hooks/use-dashboard";
import { useCountUp } from "@/hooks/use-count-up";

function StatSkeleton() {
  return <div className="h-7 w-10 bg-indigo-100 rounded-md animate-pulse" />;
}

function StatValue({ target, isLoading }: { target: number | undefined; isLoading: boolean }) {
  const display = useCountUp(target ?? 0);
  if (isLoading) return <StatSkeleton />;
  return <p className="text-xl md:text-2xl font-semibold text-gray-700 leading-none">{target != null ? display : "—"}</p>;
}

// `min-w-0` lets a label wrap on a very narrow phone instead of pushing the page wider than the screen.
const STAT_CLASS = "flex flex-col items-center justify-center gap-0.5 min-w-0";
const LABEL_CLASS =
  "text-[10px] font-semibold text-gray-400 uppercase tracking-wide md:tracking-widest flex items-center justify-center gap-1 text-center";

/**
 * A divider as its own flex item (not a border on a stat), so `justify-between` puts equal space on
 * both sides of it and it sits halfway between the stats.
 */
function Divider() {
  return <div aria-hidden className="w-0.5 self-stretch bg-indigo-300 shrink-0" />;
}

export default function DashboardStatCards() {
  const { data, isLoading } = useDashboardCardData();

  return (
    // Mobile: spread across the full width. From `md`: sized to content, next to the date picker.
    <div className="flex items-center justify-between gap-3 w-full md:w-auto md:justify-start md:gap-8 md:pl-6">
      <div className={STAT_CLASS}>
        <StatValue target={data?.source_count} isLoading={isLoading} />
        <p className={LABEL_CLASS}>
          Sources <InfoTooltip text="Number of job listing sources scraped" className="text-zinc-400 hover:text-zinc-600" />
        </p>
      </div>
      <Divider />
      <div className={STAT_CLASS}>
        <StatValue target={data?.all_time_vacancy_count} isLoading={isLoading} />
        <p className={LABEL_CLASS}>
          <span>Total Vacancies</span> <InfoTooltip text="Total number of job vacancies collected" className="text-zinc-400 hover:text-zinc-600" />
        </p>
      </div>
      <Divider />
      <div className={STAT_CLASS}>
        <StatValue target={data?.industry_count} isLoading={isLoading} />
        <p className={LABEL_CLASS}>
          Industries <InfoTooltip text="Number of industries classified under SLSIC" className="text-zinc-400 hover:text-zinc-600" />
        </p>
      </div>
    </div>
  );
}
