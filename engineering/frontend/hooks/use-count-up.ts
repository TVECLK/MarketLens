"use client";

import { useEffect, useRef, useState } from "react";

// Drops to 1 decimal place without rounding up — e.g. 10.89 → 10.8, never 10.9.
// Plain toFixed(1) would round, which makes the animated number jump ahead of
// what has actually been counted to.
function floorOneDecimal(n: number) {
  return Math.floor(n * 10) / 10;
}

// Returns the display string for the current animated value.
// target is passed so we know (a) whether to use the k suffix at all,
// and (b) whether to append "+" when the animation finishes on a non-round number.
function formatStat(value: number, target: number): string {
  // Small numbers (sources, industry count) — show as a plain integer
  if (target < 1000) return Math.round(value).toString();

  const isExactThousand = target % 1000 === 0; // e.g. 10000, 5000 — no "+" needed
  const isFinished = value >= target;           // animation has reached the target
  const k = floorOneDecimal(value / 1000).toFixed(1); // e.g. "203.8"

  // Append "+" only at the very end, and only when the number isn't a clean multiple
  // of 1000 — indicates there are vacancies beyond what the rounded figure shows
  return `${k}k${isFinished && !isExactThousand ? "+" : ""}`;
}

// Cubic ease-out curve: t goes 0 → 1, output goes 0 → 1.
// The cube makes it accelerate fast at the start and slow sharply near the end,
// so the number shoots up quickly then settles onto the final value.
const easeOut = (t: number) => 1 - Math.pow(1 - t, 3);

// Animates a number from 0 to `target` over `duration` ms and returns the
// formatted display string on every frame.
export function useCountUp(target: number, duration = 1000): string {
  const [display, setDisplay] = useState(() => formatStat(0, target));

  // Holds the rAF id so we can cancel the loop on cleanup or when target changes
  const rafRef = useRef<number | null>(null);

  // Holds the timestamp of the first animation frame — set lazily on first tick
  // so the timer starts exactly when the browser is ready, not when useEffect runs
  const startTime = useRef<number | null>(null);

  useEffect(() => {
    if (target === 0) return;

    // Reset start time so a new animation begins fresh each time target changes
    startTime.current = null;

    const tick = (now: number) => {
      // Record the start timestamp on the very first frame
      if (startTime.current === null) startTime.current = now;

      // progress: 0 at start, 1 when duration has elapsed — clamped so it never exceeds 1
      const progress = Math.min((now - startTime.current) / duration, 1);

      // Apply easing then scale to the target range to get the current display value
      const current = easeOut(progress) * target;

      setDisplay(formatStat(current, target));

      // Keep scheduling frames until the animation is complete
      if (progress < 1) rafRef.current = requestAnimationFrame(tick);
    };

    rafRef.current = requestAnimationFrame(tick);

    // Cancel any in-flight frame when the component unmounts or target changes
    return () => { if (rafRef.current) cancelAnimationFrame(rafRef.current); };
  }, [target, duration]);

  return display;
}
