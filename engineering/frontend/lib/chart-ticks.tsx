"use client";

import { useState } from "react";
import { HoverTooltip, type TooltipPos } from "@/lib/hover-tooltip";

interface TickProps {
  x?: number | string;
  y?: number | string;
  payload?: { value: string };
  maxLength: number;
  fontSize?: number;
  fill?: string;
  [key: string]: unknown;
}

function truncate(value: string, maxLength: number) {
  return value.length > maxLength ? `${value.substring(0, maxLength)}…` : value;
}

/** YAxis tick — truncates to maxLength, shows full-text portal tooltip on hover/focus. */
export function YAxisTooltipTick({ x = 0, y = 0, payload, maxLength, fontSize = 10, fill = "#a1a1aa" }: TickProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);
  const full = payload?.value ?? "";
  const label = truncate(full, maxLength);
  const isTruncated = full !== label;

  const show = (e: React.SyntheticEvent) => {
    if (!isTruncated) return;
    const rect = (e.currentTarget as Element).getBoundingClientRect();
    setPos({ mouseX: rect.left + rect.width / 2, mouseY: rect.top });
  };

  return (
    <g
      tabIndex={isTruncated ? 0 : undefined}
      aria-label={isTruncated ? full : undefined}
      onMouseEnter={(e) => isTruncated && setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
      onFocus={show}
      onBlur={() => setPos(null)}
    >
      <text x={Number(x)} y={Number(y)} dy="0.355em" textAnchor="end" fill={fill} fontSize={fontSize}>
        {label}
      </text>
      {pos && <HoverTooltip text={full} pos={pos} />}
    </g>
  );
}

/** XAxis tick — truncates to maxLength, straight (no rotation), shows full-text portal tooltip on hover/focus. */
export function XAxisStraightTooltipTick({ x = 0, y = 0, payload, maxLength, fontSize = 10, fill = "#a1a1aa" }: TickProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);
  const full = payload?.value ?? "";
  const label = truncate(full, maxLength);
  const isTruncated = full !== label;

  const show = (e: React.SyntheticEvent) => {
    if (!isTruncated) return;
    const rect = (e.currentTarget as Element).getBoundingClientRect();
    setPos({ mouseX: rect.left + rect.width / 2, mouseY: rect.top });
  };

  return (
    <g
      transform={`translate(${Number(x)},${Number(y)})`}
      tabIndex={isTruncated ? 0 : undefined}
      aria-label={isTruncated ? full : undefined}
      onMouseEnter={(e) => isTruncated && setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
      onFocus={show}
      onBlur={() => setPos(null)}
    >
      <text x={0} y={0} dy={12} textAnchor="middle" fill={fill} fontSize={fontSize}>
        {label}
      </text>
      {pos && <HoverTooltip text={full} pos={pos} />}
    </g>
  );
}

/** XAxis tick — splits label into two lines at the nearest space to the midpoint, rotated -45°. */
export function XAxisWrappedTick({ x = 0, y = 0, payload, fontSize = 9, fill = "#a1a1aa" }: TickProps) {
  const full = payload?.value ?? "";
  const mid = Math.floor(full.length / 2);
  // find the space closest to the middle
  let splitAt = full.indexOf(" ", mid);
  if (splitAt === -1) splitAt = full.lastIndexOf(" ", mid);
  const line1 = splitAt !== -1 ? full.slice(0, splitAt) : full;
  const line2 = splitAt !== -1 ? full.slice(splitAt + 1) : "";

  return (
    <g transform={`translate(${Number(x)},${Number(y)})`}>
      <text x={0} y={0} textAnchor="end" fill={fill} fontSize={fontSize} transform="rotate(-45)">
        <tspan x={0} dy={4}>{line1}</tspan>
        {line2 && <tspan x={0} dy={fontSize + 2}>{line2}</tspan>}
      </text>
    </g>
  );
}

/** XAxis tick — truncates to maxLength, rotated -45°, shows full-text portal tooltip on hover/focus. */
export function XAxisTooltipTick({ x = 0, y = 0, payload, maxLength, fontSize = 9, fill = "#a1a1aa" }: TickProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);
  const full = payload?.value ?? "";
  const label = truncate(full, maxLength);
  const isTruncated = full !== label;

  const show = (e: React.SyntheticEvent) => {
    if (!isTruncated) return;
    const rect = (e.currentTarget as Element).getBoundingClientRect();
    setPos({ mouseX: rect.left + rect.width / 2, mouseY: rect.top });
  };

  return (
    <g
      transform={`translate(${Number(x)},${Number(y)})`}
      tabIndex={isTruncated ? 0 : undefined}
      aria-label={isTruncated ? full : undefined}
      onMouseEnter={(e) => isTruncated && setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
      onFocus={show}
      onBlur={() => setPos(null)}
    >
      <text x={0} y={0} dy={4} textAnchor="end" fill={fill} fontSize={fontSize} transform="rotate(-45)">
        {label}
      </text>
      {pos && <HoverTooltip text={full} pos={pos} />}
    </g>
  );
}
