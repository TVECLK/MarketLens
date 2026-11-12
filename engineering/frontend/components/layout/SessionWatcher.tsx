"use client";

import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

/** At most one re-check in this window, since focus and visibility usually fire together. */
const MIN_INTERVAL_MS = 2000;

/**
 * Re-checks the session with the server whenever this tab becomes active again.
 *
 * Sign-in state is worked out on the server when a page renders, and the session cookie is httpOnly,
 * so a tab can't see a sign-out done in another tab. `router.refresh()` re-runs the proxy and the
 * root layout with the current cookies: on an admin page the proxy sends a signed-out user home, and
 * elsewhere the sidebar and header switch to signed out. Client state (open dialogs, cached queries)
 * is kept. Rendered only while signed in.
 */
export default function SessionWatcher() {
  const router = useRouter();
  const lastCheck = useRef(0);

  useEffect(() => {
    const recheck = () => {
      if (document.visibilityState !== "visible") return;
      const now = Date.now();
      if (now - lastCheck.current < MIN_INTERVAL_MS) return;
      lastCheck.current = now;
      router.refresh();
    };

    document.addEventListener("visibilitychange", recheck);
    window.addEventListener("focus", recheck);
    return () => {
      document.removeEventListener("visibilitychange", recheck);
      window.removeEventListener("focus", recheck);
    };
  }, [router]);

  return null;
}
