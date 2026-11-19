"use client";

import { createPortal } from "react-dom";

export interface TooltipPos { mouseX: number; mouseY: number }

export function HoverTooltip({ text, pos }: { text: string; pos: TooltipPos }) {
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
