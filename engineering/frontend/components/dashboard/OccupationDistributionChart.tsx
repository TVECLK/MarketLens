"use client";

import Link from "next/link";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts";
import { SignedOut } from "@thunderid/nextjs";
import ChartCard from "@/components/shared/ChartCard";
import { isEmptyChartData } from "@/lib/chart-data";
import {
  CHART_COLORS,
  TICK_COLOR,
  CHART_TOOLTIP_STYLE,
  CHART_TOOLTIP_ITEM_STYLE,
  CHART_TOOLTIP_LABEL_STYLE,
  CHART_BAR_CURSOR,
} from "@/lib/chart-theme";
import type { ChartDatum } from "@/types/chart";

const MAX_LABEL_LENGTH = 22;

interface OccupationDistributionChartProps {
  data: ChartDatum[];
  isLoading: boolean;
  isError: boolean;
  linkHref: string;
}

export default function OccupationDistributionChart({
  data,
  isLoading,
  isError,
  linkHref,
}: OccupationDistributionChartProps) {
  const count = data.length > 0 ? data.length : null;
  const subtitle = count ? `All ${count} standard occupations` : "Standard occupations";

  const topOccupation = data.length > 0
    ? data.reduce((max, d) => d.value > max.value ? d : max, data[0])
    : null;
  const totalVacancies = data.reduce((sum, d) => sum + d.value, 0);
  const topPercent = topOccupation && totalVacancies > 0
    ? Math.round((topOccupation.value / totalVacancies) * 100)
    : null;

  const headline = count && topOccupation && topPercent ? (
    <>
      <span className="bg-indigo-200 text-indigo-800 font-semibold px-1">{topOccupation.name}</span>
      {" "}leads demand with{" "}
      <span className="bg-indigo-200 text-indigo-800 font-semibold px-1">{topPercent}%</span>
      {" "}of vacancies classified under the Sri Lanka Standard Classification of Occupations (SLSCO).
    </>
  ) : undefined;

  const sharedProps = {
    title: "Job Distribution by Occupation (SLSCO)",
    subtitle,
    heightClassName: "h-[420px] md:h-[540px]",
    isLoading,
    isError,
    isEmpty: isEmptyChartData(data, (d) => d.value),
    loadingLabel: "Loading occupation data...",
    footer: (
      <div className="pt-8 flex justify-end">
        <Link
          href={linkHref}
          className="text-xs font-bold text-indigo-600 hover:text-indigo-800 hover:underline inline-flex items-center gap-1 transition-colors"
        >
          See More →
        </Link>
      </div>
    ),
  };

  const chart = (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={data} layout="vertical" margin={{ left: 20, right: 20, bottom: 12 }}>
        <XAxis
          type="number"
          tick={{ fontSize: 10, fill: TICK_COLOR }}
          label={{
            value: "No. of Vacancies",
            position: "insideBottom",
            offset: -8,
            fontSize: 11,
            fill: TICK_COLOR,
          }}
        />
        <YAxis
          dataKey="name"
          type="category"
          tick={{ fontSize: 10, fill: TICK_COLOR }}
          width="auto"
          tickFormatter={(v) =>
            v.length > MAX_LABEL_LENGTH ? `${v.substring(0, MAX_LABEL_LENGTH)}...` : v
          }
          label={{
            value: "Occupation",
            angle: -90,
            position: "insideLeft",
            fontSize: 11,
            fill: TICK_COLOR,
          }}
        />
        <Tooltip
          contentStyle={CHART_TOOLTIP_STYLE}
          itemStyle={CHART_TOOLTIP_ITEM_STYLE}
          labelStyle={CHART_TOOLTIP_LABEL_STYLE}
          cursor={CHART_BAR_CURSOR}
        />
        <Bar dataKey="value" name="Job Count" fill={CHART_COLORS.indigo} radius={[0, 4, 4, 0]} barSize={16} />
      </BarChart>
    </ResponsiveContainer>
  );

  return (
    <SignedOut fallback={<ChartCard {...sharedProps}>{chart}</ChartCard>}>
      <ChartCard {...sharedProps} headline={headline}>{chart}</ChartCard>
    </SignedOut>
  );
}
