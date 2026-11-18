"use client";

import { useState } from "react";
import { createPortal } from "react-dom";

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

interface TooltipPos { mouseX: number; mouseY: number }

function HoverTooltip({ text, pos }: { text: string; pos: TooltipPos }) {
  const approxWidth = text.length * 6.5 + 16;
  const centeredLeft = pos.mouseX + window.scrollX - approxWidth / 2;
  const left = Math.max(8, Math.min(centeredLeft, window.innerWidth + window.scrollX - approxWidth - 8));
  const top = pos.mouseY + window.scrollY - 34;
  return createPortal(
    <div
      style={{
        position: "absolute",
        top,
        left,
        background: "#18181b",
        color: "#fff",
        fontSize: 11,
        padding: "3px 8px",
        borderRadius: 4,
        whiteSpace: "nowrap",
        pointerEvents: "none",
        zIndex: 9999,
        boxShadow: "0 2px 8px rgba(0,0,0,0.25)",
      }}
    >
      {text}
    </div>,
    document.body,
  );
}

/** YAxis tick — truncates to maxLength, shows full-text portal tooltip on hover. */
export function YAxisTooltipTick({ x = 0, y = 0, payload, maxLength, fontSize = 10, fill = "#a1a1aa" }: TickProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);
  const full = payload?.value ?? "";
  const label = truncate(full, maxLength);
  const isTruncated = full !== label;
  return (
    <g
      onMouseEnter={(e) => isTruncated && setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
    >
      <text x={Number(x)} y={Number(y)} dy="0.355em" textAnchor="end" fill={fill} fontSize={fontSize}>
        {label}
      </text>
      {pos && <HoverTooltip text={full} pos={pos} />}
    </g>
  );
}

/** XAxis tick — truncates to maxLength, straight (no rotation), shows full-text portal tooltip on hover. */
export function XAxisStraightTooltipTick({ x = 0, y = 0, payload, maxLength, fontSize = 10, fill = "#a1a1aa" }: TickProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);
  const full = payload?.value ?? "";
  const label = truncate(full, maxLength);
  const isTruncated = full !== label;
  return (
    <g
      transform={`translate(${Number(x)},${Number(y)})`}
      onMouseEnter={(e) => isTruncated && setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
    >
      <text x={0} y={0} dy={12} textAnchor="middle" fill={fill} fontSize={fontSize}>
        {label}
      </text>
      {pos && <HoverTooltip text={full} pos={pos} />}
    </g>
  );
}

/** XAxis tick — truncates to maxLength, rotated -45°, shows full-text portal tooltip on hover. */
export function XAxisTooltipTick({ x = 0, y = 0, payload, maxLength, fontSize = 9, fill = "#a1a1aa" }: TickProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);
  const full = payload?.value ?? "";
  const label = truncate(full, maxLength);
  const isTruncated = full !== label;
  return (
    <g
      transform={`translate(${Number(x)},${Number(y)})`}
      onMouseEnter={(e) => isTruncated && setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
    >
      <text x={0} y={0} dy={4} textAnchor="end" fill={fill} fontSize={fontSize} transform="rotate(-45)">
        {label}
      </text>
      {pos && <HoverTooltip text={full} pos={pos} />}
    </g>
  );
}
