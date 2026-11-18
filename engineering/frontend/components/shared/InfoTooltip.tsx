"use client";

import { useState } from "react";
import { createPortal } from "react-dom";
import { Info } from "lucide-react";

interface InfoTooltipProps {
  text: string;
  className?: string;
}

function HoverTooltip({ text, mouseX, mouseY }: { text: string; mouseX: number; mouseY: number }) {
  const maxWidth = 240;
  const centeredLeft = mouseX + window.scrollX - maxWidth / 2;
  const left = Math.max(8, Math.min(centeredLeft, window.innerWidth + window.scrollX - maxWidth - 8));
  const top = mouseY + window.scrollY - 38;
  return createPortal(
    <div
      style={{
        position: "absolute",
        top,
        left,
        maxWidth,
        background: "#18181b",
        color: "#fff",
        fontSize: 11,
        padding: "5px 10px",
        borderRadius: 4,
        whiteSpace: "normal",
        lineHeight: 1.5,
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

export default function InfoTooltip({ text, className }: InfoTooltipProps) {
  const [pos, setPos] = useState<{ mouseX: number; mouseY: number } | null>(null);

  return (
    <span
      className={className}
      onMouseEnter={(e) => setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseMove={(e) => setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
      style={{ display: "inline-flex", alignItems: "center", cursor: "default" }}
    >
      <Info size={13} strokeWidth={2} />
      {pos && <HoverTooltip text={text} mouseX={pos.mouseX} mouseY={pos.mouseY} />}
    </span>
  );
}
