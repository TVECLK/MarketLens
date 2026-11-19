"use client";

import InfoTooltip from "@/components/shared/InfoTooltip";
import { useDashboardCardData } from "@/hooks/use-dashboard";

export default function DashboardStatCards() {
  const { data } = useDashboardCardData();

  return (
    <div className="flex items-center pl-6">
      <div className="flex flex-col items-center justify-center gap-0.5 pr-8">
        <p className="text-2xl font-semibold text-gray-700 leading-none">{data?.source_count ?? "—"}</p>
        <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest flex items-center gap-1">
          Sources <InfoTooltip text="Number of job listing sources scraped" className="text-zinc-400 hover:text-zinc-600" />
        </p>
      </div>
      <div className="flex flex-col items-center justify-center gap-0.5 border-l-2 border-indigo-300 px-8">
        <p className="text-2xl font-semibold text-gray-700 leading-none">
          {data?.all_time_vacancy_count != null ? data.all_time_vacancy_count.toLocaleString() : "—"}
        </p>
        <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest flex items-center gap-1">
          <span>Total Vacancies</span> <InfoTooltip text="Total job vacancies collected across all sources all time" className="text-zinc-400 hover:text-zinc-600" />
        </p>
      </div>
      <div className="flex flex-col items-center justify-center gap-0.5 border-l-2 border-indigo-300 pl-8">
        <p className="text-2xl font-semibold text-gray-700 leading-none">{data?.industry_count ?? "—"}</p>
        <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest flex items-center gap-1">
          Industries <InfoTooltip text="Number of industries classified under SLSIC" className="text-zinc-400 hover:text-zinc-600" />
        </p>
      </div>
    </div>
  );
}
