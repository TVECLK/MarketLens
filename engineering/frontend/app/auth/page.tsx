"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

/** How long to wait before offering a way out, in case the login can't complete. */
const SLOW_LOGIN_MS = 15000;

/**
 * Where Thunder sends the browser after the login form (`NEXT_PUBLIC_THUNDERID_AFTER_SIGN_IN_URL`).
 * The ThunderID provider completes the login from the URL's `code` while this loading screen shows;
 * the proxy then sends the user to the admin area, or signs them out if they have no access.
 */
export default function AuthPage() {
  const [slow, setSlow] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setSlow(true), SLOW_LOGIN_MS);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50" role="status" aria-live="polite">
      <div className="flex flex-col items-center gap-3 max-w-sm text-center px-4">
        <div className="w-8 h-8 rounded-full border-2 border-gray-200 border-t-gray-800 animate-spin" />
        <p className="text-sm font-bold text-gray-800">Signing you in…</p>
        {slow && (
          <p className="text-xs font-semibold text-gray-500">
            This is taking longer than expected.{" "}
            <Link href="/" className="underline">
              Back to home
            </Link>
          </p>
        )}
      </div>
    </div>
  );
}
