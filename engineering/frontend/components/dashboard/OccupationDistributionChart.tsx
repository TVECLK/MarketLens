"use client";

import Link from "next/link";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts";
import ChartCard from "@/components/shared/ChartCard";
import {
  CHART_COLORS,
  TICK_COLOR,
  CHART_TOOLTIP_STYLE,
  CHART_TOOLTIP_ITEM_STYLE,
  CHART_TOOLTIP_LABEL_STYLE,
  CHART_BAR_CURSOR,
} from "@/lib/chart-theme";
import type { ChartDatum } from "@/types/chart";

interface OccupationDistributionChartProps {
  data: ChartDatum[];
  isMobile: boolean;
  isLoading: boolean;
  linkHref: string;
}

export default function OccupationDistributionChart({
  data,
  isMobile,
  isLoading,
  linkHref,
}: OccupationDistributionChartProps) {
  return (
    <ChartCard
      title="Job Distribution by Occupation (SLSO)"
      subtitle="All 10 standard occupations"
      heightClassName="h-[420px] md:h-[540px]"
      isLoading={isLoading}
      loadingLabel="Loading occupation data..."
      footer={
        <div className="pt-8 flex justify-end">
          <Link
            href={linkHref}
            className="text-xs font-bold text-indigo-600 hover:text-indigo-800 hover:underline inline-flex items-center gap-1 transition-colors"
          >
            See More →
          </Link>
        </div>
      }
    >
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
            width={isMobile ? 90 : 170}
            tickFormatter={(v) =>
              v.length > (isMobile ? 12 : 22) ? `${v.substring(0, 22)}...` : v
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
          <Bar dataKey="value" fill={CHART_COLORS.indigo} radius={[0, 4, 4, 0]} barSize={16} />
        </BarChart>
      </ResponsiveContainer>
    </ChartCard>
  );
}
