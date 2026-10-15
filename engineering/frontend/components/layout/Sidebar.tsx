"use client";

import { SignedIn } from "@thunderid/nextjs";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { clsx } from "clsx";
import { useEffect, useRef, useState } from "react";
import Image from "next/image";

const navigation = [
  { name: "Dashboard", href: "/", icon: LayoutDashboardIcon },
  { name: "KPIs", href: "/kpis", icon: GaugeIcon, protected: true },
];

const COLLAPSED_W = "4rem"; // w-16

export default function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);
  const asideRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const content =
      (document.getElementById("main-content") as HTMLElement | null) ??
      (document.querySelector("main") as HTMLElement | null) ??
      (asideRef.current?.nextElementSibling as HTMLElement | null);
    if (!content) return;

    const computed = window.getComputedStyle(content);
    const prop: "marginLeft" | "paddingLeft" =
      parseFloat(computed.marginLeft) >= 100 ? "marginLeft" : "paddingLeft";

    content.style.transition =
      "margin-left 200ms ease, padding-left 200ms ease";
    content.style[prop] = collapsed ? COLLAPSED_W : "";

    return () => {
      content.style[prop] = "";
    };
  }, [collapsed]);

  return (
    <aside
      ref={asideRef}
      className={clsx(
        "fixed inset-y-0 left-0 z-50 bg-zinc-950 border-r border-zinc-800 flex flex-col font-sans transition-all duration-200",
        collapsed ? "w-16" : "w-64",
      )}
    >
      {/* Collapse toggle — sits on the sidebar's right border */}
      <button
        onClick={() => setCollapsed((c) => !c)}
        aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        className="absolute -right-3 top-16 z-50 w-6 h-6 rounded-full bg-zinc-900 border border-zinc-700 flex items-center justify-center text-zinc-400 hover:text-white hover:border-zinc-500 transition-colors shadow-md cursor-pointer"
      >
        <ChevronLeftIcon
          className={clsx(
            "w-3.5 h-3.5 transition-transform duration-200",
            collapsed && "rotate-180",
          )}
        />
      </button>

      {/* Logo Component */}
      <div
        className={clsx(
          "flex items-center gap-3 py-5 border-b border-zinc-800",
          collapsed ? "px-3.5 justify-center" : "px-6",
        )}
      >
        <div className="w-9 h-9 rounded-md bg-white flex items-center justify-center overflow-hidden shrink-0">
          <Image
            src="/logo.png"
            alt="TVEC Logo"
            width={45}
            height={45}
            className="object-contain scale-110"
          />
        </div>
        {!collapsed && (
          <div className="min-w-0">
            <h1 className="text-white font-bold text-sm tracking-tight leading-tight">
              TVEC
            </h1>
            <p className="text-zinc-400 text-[10px] uppercase font-semibold tracking-wider leading-tight mt-0.5">
              Labour Intelligence
            </p>
          </div>
        )}
      </div>

      {/* Navigation Links */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navigation.map((item) => {
          const isActive =
            pathname === item.href ||
            (item.href !== "/" && pathname.startsWith(item.href));

          const link = (
            <Link
              href={item.href}
              title={collapsed ? item.name : undefined}
              className={clsx(
                "flex items-center gap-3 rounded-lg text-sm font-medium transition-all duration-150",
                collapsed ? "justify-center px-0 py-2.5" : "px-3 py-2.5",
                isActive
                  ? "bg-white text-black font-semibold shadow-sm"
                  : "text-zinc-400 hover:bg-zinc-900 hover:text-white",
              )}
            >
              <item.icon className="w-4 h-4 flex-shrink-0" />
              {!collapsed && item.name}
            </Link>
          );

          return item.protected ? (
            <SignedIn key={item.name}>{link}</SignedIn>
          ) : (
            <span key={item.name}>{link}</span>
          );
        })}
      </nav>
    </aside>
  );
}

// --- Inline Minimal SVG Icon components ---

function ChevronLeftIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="15 18 9 12 15 6" />
    </svg>
  );
}

function LayoutDashboardIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
  );
}

function GaugeIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="m12 14 4-4" />
      <path d="M3.34 19a10 10 0 1 1 17.32 0" />
    </svg>
  );
}
