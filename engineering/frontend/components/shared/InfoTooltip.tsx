"use client";

import { useState } from "react";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
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
      <InformationCircleIcon className="w-[13px] h-[13px]" />
      {pos && <HoverTooltip text={text} pos={pos} />}
    </span>
  );
}
