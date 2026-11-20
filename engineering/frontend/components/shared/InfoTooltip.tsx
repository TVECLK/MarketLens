"use client";

import { useId, useState, type FocusEvent, type MouseEvent } from "react";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
import { HoverTooltip, type TooltipPos } from "@/lib/hover-tooltip";

interface InfoTooltipProps {
  text: string;
  className?: string;
}

/** Anchor the tooltip to the icon itself (centred above it), so it lands in the same place whether
 *  it was opened by the mouse or the keyboard. */
function anchorPos(el: HTMLElement): TooltipPos {
  const r = el.getBoundingClientRect();
  return { mouseX: r.left + r.width / 2, mouseY: r.top };
}

/**
 * An ⓘ icon that explains something. Opens on hover and on keyboard focus (Tab), closes on mouse
 * leave, blur or Escape. Screen readers get the explanation as the button's description.
 */
export default function InfoTooltip({ text, className }: InfoTooltipProps) {
  const [pos, setPos] = useState<TooltipPos | null>(null);
  const descriptionId = useId();

  const open = (e: MouseEvent<HTMLButtonElement> | FocusEvent<HTMLButtonElement>) =>
    setPos(anchorPos(e.currentTarget));

  return (
    <button
      type="button"
      aria-label="More information"
      aria-describedby={descriptionId}
      onMouseEnter={open}
      onMouseLeave={() => setPos(null)}
      onFocus={open}
      onBlur={() => setPos(null)}
      onKeyDown={(e) => e.key === "Escape" && setPos(null)}
      className={`inline-flex items-center rounded-full cursor-default outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 focus-visible:ring-offset-1 ${className ?? ""}`}
    >
      <InformationCircleIcon aria-hidden className="w-[13px] h-[13px]" />
      {/* Always in the page (visually hidden), so screen readers can read it as soon as the button
          gets focus, without waiting for the visual tooltip to render. */}
      <span id={descriptionId} className="sr-only">
        {text}
      </span>
      {pos && <HoverTooltip text={text} pos={pos} />}
    </button>
  );
}
