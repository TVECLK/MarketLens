"use client";

import Link from "next/link";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts";
import { SignedIn, SignedOut } from "@thunderid/nextjs";
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
import { XAxisTooltipTick } from "@/lib/chart-ticks";
import type { ChartDatum } from "@/types/chart";

interface IndustryDistributionChartProps {
  data: ChartDatum[];
  isLoading: boolean;
  isError: boolean;
  linkHref: string;
}

export default function IndustryDistributionChart({
  data,
  isLoading,
  isError,
  linkHref,
}: IndustryDistributionChartProps) {
  const subtitle = data.length > 0 ? `All ${data.length} standard industries` : "Standard industries";

  const topIndustry = data.length > 0
    ? data.reduce((max, d) => d.value > max.value ? d : max, data[0])
    : null;

  const headline = topIndustry && topIndustry.value > 0 ? (
    <>
      <span className="bg-yellow-200 text-yellow-950 font-semibold px-1">{topIndustry.name}</span>
      {" "}have the highest job count according to the Sri Lanka Standard Industrial Classification (SLSIC).
    </>
  ) : undefined;

  const sharedProps = {
    title: "Job Distribution by Industry (SLSIC)",
    subtitle,
    heightClassName: "h-[480px] md:h-[640px]",
    bodyClassName: "pb-10",
    isLoading,
    isError,
    isEmpty: isEmptyChartData(data, (d) => d.value),
    loadingLabel: "Loading industry data...",
    footer: (
      <div className="pt-2 flex justify-end">
        <Link
          href={linkHref}
          className="text-sm font-bold text-teal-600 hover:text-teal-800 hover:underline inline-flex items-center gap-1 transition-colors"
        >
          See More →
        </Link>
      </div>
    ),
  };

  const chart = (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={data} margin={{ left: 30, bottom: 14 }}>
        <XAxis
          dataKey="name"
          tick={(props) => <XAxisTooltipTick {...props} maxLength={16} fontSize={9} fill={TICK_COLOR} />}
          interval={0}
          height={120}
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
          allowDecimals={false}
          label={{
            value: "No. of Vacancies",
            angle: -90,
            position: "center",
            dx: -30,
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
        <Bar dataKey="value" name="Job Count" fill={CHART_COLORS.teal} radius={[4, 4, 0, 0]} barSize={22} />
      </BarChart>
    </ResponsiveContainer>
  );

  return (
    <>
      <SignedIn>
        <ChartCard {...sharedProps}>{chart}</ChartCard>
      </SignedIn>
      <SignedOut>
        <ChartCard {...sharedProps} heightClassName="h-[620px] md:h-[740px]" headline={headline}>{chart}</ChartCard>
      </SignedOut>
    </>
  );
}
