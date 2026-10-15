"use client";

import {
  SignedIn,
  SignedOut,
  SignInButton,
  UserDropdown,
} from "@thunderid/nextjs";
import Image from "next/image";

interface HeaderProps {
  title: string;
  subtitle?: string;
  logoSrc?: string;
}

export default function Header({
  title,
  subtitle,
  logoSrc,
}: HeaderProps) {
  return (
    <header className="bg-white border-b border-gray-100 px-4 py-3 flex flex-col md:flex-row justify-between items-start md:items-center gap-2 sticky top-0 z-40">
      {/* Left: Logo (optional, signed-out only) + Page Title */}
      <div className="flex items-center gap-3">
        {logoSrc && (
          <SignedOut>
            <Image
              src={logoSrc}
              alt=""
              width={80}
              height={80}
              className="rounded-md shrink-0"
            />
          </SignedOut>
        )}
        <div>
          <h1 className="text-xl font-black text-gray-900 tracking-tight">
            {title}
          </h1>
          {subtitle && (
            <p className="text-xs text-gray-400 mt-1 font-medium">{subtitle}</p>
          )}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-4 ml-auto md:ml-0 w-full md:w-auto justify-end">
        <SignedIn>
          <UserDropdown />
        </SignedIn>
        <SignedOut>
          <SignInButton>
            {({
              signIn,
              isLoading,
            }: {
              signIn?: () => Promise<void>;
              isLoading?: boolean;
            }) => (
              <button
                className="bg-black text-white px-4 py-2 rounded-md text-sm font-medium cursor-pointer hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
                onClick={() => void signIn?.()}
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
