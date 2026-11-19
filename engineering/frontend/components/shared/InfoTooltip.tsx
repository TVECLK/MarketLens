"use client";

import { useState } from "react";
import { Info } from "lucide-react";
import { HoverTooltip, type TooltipPos } from "@/lib/hover-tooltip";

interface InfoTooltipProps {
  text: string;
  className?: string;
}

export default function InfoTooltip({ text, className }: InfoTooltipProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);

  return (
    <span
      className={className}
      onMouseEnter={(e) => setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseMove={(e) => setPos({ mouseX: e.clientX, mouseY: e.clientY })}
      onMouseLeave={() => setPos(null)}
      style={{ display: "inline-flex", alignItems: "center", cursor: "default" }}
    >
      <Info size={13} strokeWidth={2} />
      {pos && <HoverTooltip text={text} pos={pos} />}
    </span>
  );
}
