"use client";

import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts";
import ChartCard from "@/components/shared/ChartCard";
import { isEmptyChartData } from "@/lib/chart-data";
import {
  CHART_COLORS,
  GRID_STROKE,
  TICK_COLOR,
  CHART_TOOLTIP_STYLE,
  CHART_TOOLTIP_ITEM_STYLE,
  CHART_TOOLTIP_LABEL_STYLE,
} from "@/lib/chart-theme";
import type { VacancyTrendPoint } from "@/types/dashboard";

interface VacancyTrendChartProps {
  data: VacancyTrendPoint[];
  totalVacancies?: number;
  fromDate: string;
  toDate: string;
  isLoading: boolean;
  isError: boolean;
}

function formatDate(dateStr: string): string {
  const [year, month, day] = dateStr.split("-");
  const date = new Date(Number(year), Number(month) - 1, Number(day));
  return date.toLocaleDateString("en-US", { day: "numeric", month: "long", year: "numeric" });
}

export default function VacancyTrendChart({
  data,
  totalVacancies,
  fromDate,
  toDate,
  isLoading,
  isError,
}: VacancyTrendChartProps) {
  const total = isLoading
    ? "…"
    : totalVacancies === undefined
      ? "—"
      : totalVacancies.toLocaleString();

  const from = formatDate(fromDate);
  const to = formatDate(toDate);

  const title = `Sri Lanka saw ${total} job vacancies between ${from} and ${to}`;

  return (
    <ChartCard
      title={title}
      heightClassName="h-[320px] md:h-[600px]"
      isLoading={isLoading}
      isError={isError}
      isEmpty={isEmptyChartData(data, (point) => point.open_job_count)}
      loadingLabel="Loading trend data..."
    >
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 10, right: 20, left: 4, bottom: 15 }}>
          <defs>
            <linearGradient id="trendFill" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={CHART_COLORS.indigo} stopOpacity={0.15} />
              <stop offset="100%" stopColor={CHART_COLORS.indigo} stopOpacity={0.01} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke={GRID_STROKE} />
          <XAxis
            dataKey="label"
            tick={{ fontSize: 10, fill: TICK_COLOR }}
            interval="preserveStartEnd"
            minTickGap={40}
            label={{
              value: "Time Period",
              position: "insideBottom",
              offset: -10,
              fontSize: 11,
              fill: TICK_COLOR,
            }}
          />
          <YAxis
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
          />
          <Area
            type="monotone"
            dataKey="open_job_count"
            name="Job Count"
            stroke={CHART_COLORS.indigo}
            strokeWidth={2}
            fill="url(#trendFill)"
            dot={
              data.length <= 40
                ? { r: 3, fill: CHART_COLORS.indigo, strokeWidth: 0 }
                : false
            }
          />
        </AreaChart>
      </ResponsiveContainer>
    </ChartCard>
  );
}
