"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useThunderID } from "@thunderid/nextjs";

const MESSAGE_MIN_VISIBLE_MS = 3000;

/**
 * Where the proxy sends a signed-in user without app access (none of the page scopes, e.g.
 * `kpis:access` or `datasource:access`). Signs them out with the
 * SDK's own sign-out, which clears this app's session and ends the Thunder session too (so the next
 * "Sign in" asks for a password again), then lands them on the public home page.
 */
export default function AccessDeniedPage() {
  const { signOut } = useThunderID();
  const router = useRouter();

  useEffect(() => {
    if (!signOut) return;

    const timer = setTimeout(async () => {
      try {
        // On success this navigates to Thunder's sign-out, which returns the browser to "/".
        const result = await signOut();
        if (result && typeof result === "object" && "redirected" in result && result.redirected) return;
      } catch {
        // Fall through: the session cookie is cleared even when Thunder's sign-out fails.
      }
      router.replace("/");
      router.refresh();
    }, MESSAGE_MIN_VISIBLE_MS);

    return () => clearTimeout(timer);
  }, [signOut, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50" role="status" aria-live="polite">
      <div className="flex flex-col items-center gap-3 max-w-sm text-center px-4">
        <div className="w-8 h-8 rounded-full border-2 border-gray-200 border-t-gray-800 animate-spin" />
        <p className="text-sm font-bold text-gray-800">Your account doesn&apos;t have access to this application.</p>
      </div>
    </div>
  );
}
