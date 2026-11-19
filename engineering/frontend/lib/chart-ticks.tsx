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
