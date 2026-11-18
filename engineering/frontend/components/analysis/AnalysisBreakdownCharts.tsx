"use client";

import {
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
  BarChart,
  Bar,
} from "recharts";
import {
  CHART_COLORS,
  TICK_COLOR,
  CHART_TOOLTIP_STYLE,
  CHART_TOOLTIP_ITEM_STYLE,
  CHART_TOOLTIP_LABEL_STYLE,
  CHART_BAR_CURSOR,
  SECTOR_COLORS,
  GENDER_COLORS,
  buildCategoricalPalette,
  buildFormalInformalPalette,
  buildProvincePalette,
  type HierarchyAccent,
} from "@/lib/chart-theme";
import { YAxisTooltipTick, XAxisStraightTooltipTick } from "@/lib/chart-ticks";
import SectionMessage from "@/components/analysis/SectionMessage";
import InfoTooltip from "@/components/shared/InfoTooltip";
import type { AnalysisSectionErrors } from "@/types/analysis";
import type { ChartDatum, ShareDatum } from "@/types/chart";

interface AnalysisBreakdownChartsProps {
  accent: HierarchyAccent;
  childrenBreakdown: ChartDatum[];
  childrenBreakdownTitle: string;
  childrenBreakdownDescription: string;
  sectorData: ChartDatum[];
  expData: ChartDatum[];
  provinceData: ChartDatum[];
  eduData: ChartDatum[];
  formalData: ChartDatum[];
  genderData: ChartDatum[];
  nvqData: ChartDatum[];
  remoteShare: ShareDatum[];
  contractShare: ShareDatum[];
  sectionErrors: AnalysisSectionErrors;
}

export default function AnalysisBreakdownCharts({
  accent,
  childrenBreakdown,
  childrenBreakdownTitle,
  childrenBreakdownDescription,
  sectorData,
  expData,
  provinceData,
  eduData,
  formalData,
  genderData,
  nvqData,
  remoteShare,
  contractShare,
  sectionErrors,
}: AnalysisBreakdownChartsProps) {
  const categoricalPalette = buildCategoricalPalette(accent);
  const formalPalette = buildFormalInformalPalette(accent);
  const provincePalette = buildProvincePalette(accent);

  return (
    <div className="p-4 md:p-5 space-y-6">
      {childrenBreakdown.length > 0 && (
        <div className="border border-zinc-100 p-5 rounded-xl flex flex-col">
          <div className="border-b border-zinc-100 pb-2">
            <h4 className="text-sm font-bold text-zinc-900 flex items-center gap-1.5">{childrenBreakdownTitle} <InfoTooltip text="Vacancy breakdown by sub-classification within the selected category" className="text-zinc-900" /></h4>
            <p className="text-[11px] text-zinc-400">{childrenBreakdownDescription}</p>
          </div>
          <div
            className="mt-4 min-w-0"
            style={{ height: Math.max(childrenBreakdown.length * 44 + 40, 160) }}
          >
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={childrenBreakdown} layout="vertical" margin={{ left: 16, right: 20, bottom: 30 }}>
                <XAxis type="number" tick={{ fontSize: 10, fill: TICK_COLOR }} allowDecimals={false} label={{ value: "No. of Vacancies", position: "insideBottom", offset: -15, fontSize: 11, fill: TICK_COLOR }} />
                <YAxis
                  dataKey="name"
                  type="category"
                  tick={(props) => <YAxisTooltipTick {...props} maxLength={26} fontSize={10} fill={TICK_COLOR} />}
                  width={180}
                  label={{ value: childrenBreakdownTitle.toLowerCase().includes("industry") ? "Industry" : "Occupation", angle: -90, position: "insideLeft", dx: -8, fontSize: 11, fill: TICK_COLOR }}
                />
                <Tooltip
                  contentStyle={CHART_TOOLTIP_STYLE}
                  itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                  labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                  cursor={CHART_BAR_CURSOR}
                />
                <Bar dataKey="value" name="Job Count" fill={accent.primary} radius={[0, 4, 4, 0]} barSize={18} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <div className="border border-zinc-100 p-5 rounded-xl h-[380px] flex flex-col">
          <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 flex items-center gap-1.5">
            Employment Sector <InfoTooltip text="Distribution of vacancies by employment sector" className="text-zinc-900" />
          </h4>
          <div className="flex-1 mt-4 min-w-0">
            {sectionErrors.employment_sector ? (
              <SectionMessage />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={sectorData} margin={{ top: 10, right: 10, left: 10, bottom: 30 }}>
                  <XAxis dataKey="name" tick={{ fontSize: 10, fill: TICK_COLOR }} label={{ value: "Sector", position: "insideBottom", offset: -15, fontSize: 11, fill: TICK_COLOR }} />
                  <YAxis tick={{ fontSize: 10, fill: TICK_COLOR }} allowDecimals={false} label={{ value: "No. of Vacancies", angle: -90, position: "center", fontSize: 11, fill: TICK_COLOR }} />
                  <Tooltip
                    contentStyle={CHART_TOOLTIP_STYLE}
                    itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                    labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                    cursor={CHART_BAR_CURSOR}
                  />
                  <Bar dataKey="value" name="Job Count" radius={[4, 4, 0, 0]} barSize={36}>
                    {sectorData.map((_, i) => (
                      <Cell key={i} fill={SECTOR_COLORS[i % SECTOR_COLORS.length]} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        <div className="border border-zinc-100 p-5 rounded-xl h-[380px] flex flex-col">
          <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 flex items-center gap-1.5">
            Experience <InfoTooltip text="Required years of experience for the vacancies" className="text-zinc-900" />
          </h4>
          <div className="flex-1 mt-4 min-w-0">
            {sectionErrors.experience ? (
              <SectionMessage />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={expData} margin={{ top: 10, right: 10, left: 10, bottom: 40 }}>
                  <XAxis dataKey="name" tick={(props) => <XAxisStraightTooltipTick {...props} maxLength={8} fontSize={9} fill={TICK_COLOR} />} interval={0} height={50} label={{ value: "Experience Level", position: "insideBottom", offset: -30, fontSize: 11, fill: TICK_COLOR }} />
                  <YAxis tick={{ fontSize: 10, fill: TICK_COLOR }} allowDecimals={false} label={{ value: "No. of Vacancies", angle: -90, position: "center", fontSize: 11, fill: TICK_COLOR }} />
                  <Tooltip
                    contentStyle={CHART_TOOLTIP_STYLE}
                    itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                    labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                    cursor={CHART_BAR_CURSOR}
                  />
                  <Bar dataKey="value" name="Job Count" fill={CHART_COLORS.sky} radius={[4, 4, 0, 0]} barSize={30} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <div className="border border-zinc-100 p-5 rounded-xl h-[380px] flex flex-col">
          <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 flex items-center gap-1.5">
            Province <InfoTooltip text="Distribution of vacancies by province" className="text-zinc-900" />
          </h4>
          <div className="flex-1 flex items-center justify-center min-w-0">
            {sectionErrors.province ? (
              <SectionMessage />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={provinceData}
                    cx="50%"
                    cy="45%"
                    innerRadius={60}
                    outerRadius={85}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {provinceData.map((_, i) => (
                      <Cell key={i} fill={provincePalette[i % provincePalette.length]} stroke="#ffffff" />
                    ))}
                  </Pie>
                  <Legend
                    verticalAlign="bottom"
                    iconSize={8}
                    iconType="circle"
                    wrapperStyle={{ fontSize: 10, color: TICK_COLOR }}
                  />
                  <Tooltip
                    contentStyle={CHART_TOOLTIP_STYLE}
                    itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                    labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        <div className="border border-zinc-100 p-5 rounded-xl h-[380px] flex flex-col">
          <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 flex items-center gap-1.5">
            Education Level <InfoTooltip text="Required education level for the vacancies" className="text-zinc-900" />
          </h4>
          <div className="flex-1 flex items-center justify-center min-w-0">
            {sectionErrors.education ? (
              <SectionMessage />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={eduData}
                    cx="50%"
                    cy="45%"
                    innerRadius={60}
                    outerRadius={85}
                    paddingAngle={3}
                    dataKey="value"
                  >
                    {eduData.map((_, i) => (
                      <Cell key={i} fill={categoricalPalette[i % categoricalPalette.length]} stroke="#ffffff" />
                    ))}
                  </Pie>
                  <Legend
                    verticalAlign="bottom"
                    iconSize={8}
                    iconType="circle"
                    wrapperStyle={{ fontSize: 11, color: TICK_COLOR }}
                  />
                  <Tooltip
                    contentStyle={CHART_TOOLTIP_STYLE}
                    itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                    labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <div className="border border-zinc-100 p-5 rounded-xl h-[380px] flex flex-col">
          <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 flex items-center gap-1.5">
            Formal / Informal <InfoTooltip text="Proportion of formal vs informal sector vacancies" className="text-zinc-900" />
          </h4>
          <div className="flex-1 flex items-center justify-center min-w-0">
            {sectionErrors.formality ? (
              <SectionMessage />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={formalData}
                    cx="50%"
                    cy="45%"
                    innerRadius={60}
                    outerRadius={85}
                    paddingAngle={3}
                    dataKey="value"
                  >
                    {formalData.map((_, i) => (
                      <Cell key={i} fill={formalPalette[i % formalPalette.length]} stroke="#ffffff" />
                    ))}
                  </Pie>
                  <Legend
                    verticalAlign="bottom"
                    iconSize={8}
                    iconType="circle"
                    wrapperStyle={{ fontSize: 11, color: TICK_COLOR }}
                  />
                  <Tooltip
                    contentStyle={CHART_TOOLTIP_STYLE}
                    itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                    labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        <div className="border border-zinc-100 p-5 rounded-xl h-[380px] flex flex-col">
          <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 flex items-center gap-1.5">
            Gender <InfoTooltip text="Gender preference specified in vacancy listings" className="text-zinc-900" />
          </h4>
          <div className="flex-1 flex items-center justify-center min-w-0">
            {sectionErrors.gender ? (
              <SectionMessage />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={genderData}
                    cx="50%"
                    cy="45%"
                    innerRadius={60}
                    outerRadius={85}
                    paddingAngle={3}
                    dataKey="value"
                  >
                    {genderData.map((_, i) => (
                      <Cell key={i} fill={GENDER_COLORS[i % GENDER_COLORS.length]} stroke="#ffffff" />
                    ))}
                  </Pie>
                  <Legend
                    verticalAlign="bottom"
                    iconSize={8}
                    iconType="circle"
                    wrapperStyle={{ fontSize: 11, color: TICK_COLOR }}
                  />
                  <Tooltip
                    contentStyle={CHART_TOOLTIP_STYLE}
                    itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                    labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>
      </div>

      <div className="border border-zinc-100 p-5 rounded-xl h-[380px] flex flex-col">
        <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2 flex items-center gap-1.5">
          Vocational Education (NVQ Level) <InfoTooltip text="Distribution of vacancies by NVQ vocational education level" className="text-zinc-900" />
        </h4>
        <div className="flex-1 mt-4 min-w-0">
          {sectionErrors.vocational_education ? (
            <SectionMessage />
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={nvqData} margin={{ top: 10, right: 10, left: 10, bottom: 30 }}>
                <XAxis dataKey="name" tick={{ fontSize: 10, fill: TICK_COLOR }} label={{ value: "NVQ Level", position: "insideBottom", offset: -15, fontSize: 11, fill: TICK_COLOR }} />
                <YAxis tick={{ fontSize: 10, fill: TICK_COLOR }} allowDecimals={false} label={{ value: "No. of Vacancies", angle: -90, position: "center", fontSize: 11, fill: TICK_COLOR }} />
                <Tooltip
                  contentStyle={CHART_TOOLTIP_STYLE}
                  itemStyle={CHART_TOOLTIP_ITEM_STYLE}
                  labelStyle={CHART_TOOLTIP_LABEL_STYLE}
                  cursor={CHART_BAR_CURSOR}
                />
                <Bar dataKey="value" name="Job Count" fill={CHART_COLORS.violet} radius={[4, 4, 0, 0]} barSize={28} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 md:gap-8">
        <div className="border border-zinc-100 p-6 rounded-xl">
          <h3 className="text-sm font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <span className="w-2 h-2 rounded-full" style={{ backgroundColor: accent.primary }} />
            Remote / On-Site / Hybrid
          </h3>
          {sectionErrors.work_modes ? (
            <SectionMessage />
          ) : (
            <div className="space-y-4">
              {remoteShare.map((item, i) => (
                <div key={i}>
                  <div className="flex justify-between text-xs mb-1 font-medium text-zinc-600">
                    <span>{item.label}</span>
                    <span className="font-mono">{item.share}%</span>
                  </div>
                  <div className="w-full h-2 bg-zinc-100 rounded-full overflow-hidden">
                    <div
                      className="h-full transition-all duration-700"
                      style={{ width: `${item.share}%`, backgroundColor: accent.primary }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="border border-zinc-100 p-6 rounded-xl">
          <h3 className="text-sm font-bold text-zinc-900 mb-4 flex items-center gap-2">
            <span className="w-2 h-2 rounded-full" style={{ backgroundColor: accent.secondary }} />
            Job Type
          </h3>
          {sectionErrors.job_type ? (
            <SectionMessage />
          ) : (
            <div className="space-y-4">
              {contractShare.map((jt, i) => (
                <div key={i}>
                  <div className="flex justify-between text-xs mb-1 font-medium text-zinc-600">
                    <span>{jt.label}</span>
                    <span className="font-mono">{jt.share}%</span>
                  </div>
                  <div className="w-full h-2 bg-zinc-100 rounded-full overflow-hidden">
                    <div
                      className="h-full transition-all duration-700"
                      style={{ width: `${jt.share}%`, backgroundColor: accent.secondary }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
