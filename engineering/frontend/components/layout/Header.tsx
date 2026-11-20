"use client";

import {
  SignedIn,
  SignedOut,
  SignInButton,
  UserDropdown,
} from "@thunderid/nextjs";
import Image, { type StaticImageData } from "next/image";
import { THEME } from "@/lib/theme";

interface HeaderProps {
  title: string;
  subtitle?: string;
  logo?: StaticImageData;
}

export default function Header({
  title,
  subtitle,
  logo,
}: HeaderProps) {
  return (
    // One row on every screen: title on the left, sign-in / user menu top-right. On a phone the title
    // wraps beside the button, so the header doesn't spend a whole line on it.
    <header className="bg-white border-b border-gray-100 px-4 py-3 flex flex-row justify-between items-start md:items-center gap-3">
      {/* Left: Logo (optional, signed-out only) + Page Title */}
      {/* `min-w-0` on this row and the text block lets long text wrap instead of overflowing the screen. */}
      <div className="flex flex-1 items-center gap-3 min-w-0">
        {logo && (
          <SignedOut>
            <Image
              src={logo}
              alt="TVEC logo"
              className="h-8 w-auto md:h-12 rounded-md shrink-0"
            />
          </SignedOut>
        )}
        <div className="min-w-0 flex-1">
          <h1 className="text-[12px] md:text-xl font-black text-gray-900 tracking-tight leading-tight text-balance break-words">
            {title}
          </h1>
          {subtitle && (
            <p className="text-[8px] md:text-xs text-gray-400 mt-0.5 md:mt-1 font-medium leading-snug text-pretty break-words">
              {subtitle}
            </p>
          )}
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-4">
        <SignedIn>
          <UserDropdown />
        </SignedIn>
        <SignedOut>
          <SignInButton>
            {({ signIn, isLoading }) => (
              <button
                className={`${THEME.button} px-3 py-1.5 md:px-4 md:py-2 rounded-md text-xs md:text-sm font-medium whitespace-nowrap cursor-pointer disabled:cursor-not-allowed disabled:opacity-60 transition-colors`}
                onClick={() => signIn().catch((err) => console.error("Sign-in failed", err))}
                disabled={isLoading}
              >
                {isLoading ? "Signing in…" : "Sign in"}
              </button>
            )}
          </SignInButton>
        </SignedOut>
      </div>
    </header>
  );
}
