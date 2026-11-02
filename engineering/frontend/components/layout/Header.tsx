"use client";

import {
  SignedIn,
  SignedOut,
  SignInButton,
  UserDropdown,
} from "@thunderid/nextjs";
import Image, { type StaticImageData } from "next/image";

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
    <header className="bg-white border-b border-gray-100 px-4 py-3 flex flex-col md:flex-row justify-between items-start md:items-center gap-2">
      {/* Left: Logo (optional, signed-out only) + Page Title */}
      <div className="flex items-center gap-3">
        {logo && (
          <SignedOut>
            <Image
              src={logo}
              alt="TVEC logo"
              className="h-8 w-auto md:h-12 rounded-md shrink-0"
            />
          </SignedOut>
        )}
        <div>
          <h1 className="text-base md:text-xl font-black text-gray-900 tracking-tight">
            {title}
          </h1>
          {subtitle && (
            <p className="text-xs text-gray-400 mt-1 font-medium">{subtitle}</p>
          )}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-4 w-full md:w-auto justify-end">
        <SignedIn>
          <UserDropdown />
        </SignedIn>
        <SignedOut>
          <SignInButton>
            {({ signIn, isLoading }) => (
              <button
                className="bg-black text-white px-4 py-2 rounded-md text-sm font-medium cursor-pointer hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
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
