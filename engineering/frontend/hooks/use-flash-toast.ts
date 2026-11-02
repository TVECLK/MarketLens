"use client";

import { useCallback, useEffect, useRef, useState } from "react";

/** A single short-lived message. A new message replaces the current one and
 * restarts the timer, so back-to-back actions don't cut each other short. */
export function useFlashToast(durationMs = 2200) {
  const [toast, setToast] = useState<string | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const flashToast = useCallback(
    (msg: string) => {
      if (timer.current) clearTimeout(timer.current);
      setToast(msg);
      timer.current = setTimeout(() => setToast(null), durationMs);
    },
    [durationMs],
  );

  useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current);
  }, []);

  return { toast, flashToast };
}
