"use client";

import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts";
import {
  CHART_COLORS,
  TICK_COLOR,
  CHART_TOOLTIP_STYLE,
  CHART_TOOLTIP_ITEM_STYLE,
  CHART_TOOLTIP_LABEL_STYLE,
  CHART_BAR_CURSOR,
} from "@/lib/chart-theme";
import type { ChartDatum } from "@/types/chart";
import type { SkillStat, EmployerStat } from "@/types/occupation";

interface SkillAnalysisPanelProps {
  isLoading: boolean;
  isError: boolean;
  topSkills: ChartDatum[];
  allSkills: SkillStat[];
  totalSkillsCount: number;
  totalSkillMentions: number;
  topEmployers: EmployerStat[];
}

export default function SkillAnalysisPanel({
  isLoading,
  isError,
  topSkills,
  allSkills,
  totalSkillsCount,
  totalSkillMentions,
  topEmployers,
}: SkillAnalysisPanelProps) {
  return (
    <div className="p-4 md:p-5 space-y-6">
      {isLoading ? (
        <div className="p-10 text-center text-xs text-zinc-400">Loading skills...</div>
      ) : isError ? (
        <div className="p-10 text-center text-xs text-red-500 font-bold">
          Failed to load skills data.
        </div>
      ) : (
        <>
          <div className="border border-zinc-100 p-5 rounded-xl flex flex-col">
            <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2">
              Top 15 Skills in Demand
            </h4>
            <div
              className="mt-4 min-w-0"
              style={{ height: Math.max(topSkills.length * 32 + 40, 200) }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={topSkills} layout="vertical" margin={{ left: 20, right: 20 }}>
                  <XAxis type="number" tick={{ fontSize: 10, fill: TICK_COLOR }} />
                  <YAxis
                    dataKey="name"
                    type="category"
                    tick={{ fontSize: 10, fill: TICK_COLOR }}
                    width={150}
                  />
                  <Tooltip
                    contentStyle={CHART_TOOLTIP_STYLE}
                    itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                    labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                    cursor={CHART_BAR_CURSOR}
                  />
                  <Bar dataKey="value" fill={CHART_COLORS.indigo} radius={[0, 4, 4, 0]} barSize={16} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
            <div className="border border-zinc-100 p-5 rounded-xl">
              <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 mb-3">
                All Skills ({allSkills.length} of {totalSkillsCount})
              </h4>
              <div className="max-h-[420px] overflow-y-auto rounded-lg border border-zinc-100">
                <table className="w-full text-left">
                  <thead className="bg-zinc-50 sticky top-0">
                    <tr className="border-b border-zinc-200">
                      <th className="px-3 py-2 text-[10px] font-black uppercase tracking-wider text-zinc-400">
                        Skill
                      </th>
                      <th className="px-3 py-2 text-[10px] font-black uppercase tracking-wider text-zinc-400 text-right">
                        Count
                      </th>
                      <th className="px-3 py-2 text-[10px] font-black uppercase tracking-wider text-zinc-400 text-right w-16">
                        Share
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {allSkills.length === 0 ? (
                      <tr>
                        <td colSpan={3} className="px-3 py-6 text-center text-xs text-zinc-400">
                          No skills recorded.
                        </td>
                      </tr>
                    ) : (
                      allSkills.map((s) => (
                        <tr
                          key={s.id}
                          className="border-b border-zinc-50 last:border-0 hover:bg-zinc-50 transition-colors"
                        >
                          <td className="px-3 py-2 text-xs font-bold text-zinc-800">{s.skill}</td>
                          <td className="px-3 py-2 text-xs font-mono font-bold text-zinc-800 text-right">
                            {s.open_job_count.toLocaleString()}
                          </td>
                          <td className="px-3 py-2 text-xs font-mono text-zinc-400 text-right">
                            {Math.round((s.open_job_count / totalSkillMentions) * 100)}%
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="border border-zinc-100 p-5 rounded-xl">
              <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 mb-3">
                Top Hiring Employers
              </h4>
              <div className="space-y-3">
                {topEmployers.length === 0 ? (
                  <p className="text-xs text-zinc-400 text-center py-6">
                    No employer data available.
                  </p>
                ) : (
                  topEmployers.map((emp, i) => (
                    <div key={emp.id} className="flex items-center gap-3">
                      <span className="w-5 h-5 rounded-full bg-zinc-900 text-white text-[10px] font-black flex items-center justify-center shrink-0">
                        {i + 1}
                      </span>
                      <span className="flex-1 text-xs font-bold text-zinc-800 truncate">
                        {emp.name}
                      </span>
                      <span className="text-xs font-mono font-bold text-zinc-500 shrink-0">
                        {emp.open_job_count.toLocaleString()}
                      </span>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
