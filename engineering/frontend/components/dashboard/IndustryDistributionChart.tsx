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

interface IndustryDistributionChartProps {
  data: ChartDatum[];
  isLoading: boolean;
  linkHref: string;
}

export default function IndustryDistributionChart({
  data,
  isLoading,
  linkHref,
}: IndustryDistributionChartProps) {
  return (
    <ChartCard
      title="Job Distribution by Industry (SLSIC)"
      subtitle="All 21 standard industries"
      heightClassName="h-[480px] md:h-[640px]"
      bodyClassName="pb-10"
      isLoading={isLoading}
      loadingLabel="Loading industry data..."
      footer={
        <div className="pt-2 flex justify-end">
          <Link
            href={linkHref}
            className="text-xs font-bold text-teal-600 hover:text-teal-800 hover:underline inline-flex items-center gap-1 transition-colors"
          >
            See More →
          </Link>
        </div>
      }
    >
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} margin={{ bottom: 14 }}>
          <XAxis
            dataKey="name"
            tick={{ fontSize: 9, fill: TICK_COLOR }}
            interval={0}
            angle={-45}
            textAnchor="end"
            height={100}
            tickFormatter={(v: string) => (v.length > 20 ? `${v.substring(0, 20)}...` : v)}
            label={{
              value: "Industry",
              position: "insideBottom",
              offset: -10,
              fontSize: 11,
              fill: TICK_COLOR,
            }}
          />
          <YAxis
            type="number"
            tick={{ fontSize: 10, fill: TICK_COLOR }}
            label={{
              value: "No. of Vacancies",
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
          <Bar dataKey="value" fill={CHART_COLORS.teal} radius={[4, 4, 0, 0]} barSize={22} />
        </BarChart>
      </ResponsiveContainer>
    </ChartCard>
  );
}
