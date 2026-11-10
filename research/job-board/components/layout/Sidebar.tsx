"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { clsx } from "clsx";
import Image from "next/image";
import { useEffect, useRef, useState } from "react";

const navigation = [
  { name: "Dashboard", href: "/", icon: LayoutDashboardIcon },
  // { name: "Vacancies", href: "/vacancies", icon: BriefcaseIcon },
  // { name: "Skill Analysis", href: "/categories", icon: BarChart3Icon },
  { name: "Forecasting", href: "/forecasting", icon: ForecastingIcon },
  { name: "KPIs", href: "/kpis", icon: GaugeIcon },
  { name: "Crawling Details", href: "/crawling-details", icon: ScanSearchIcon },
];

const EXPANDED_W = "16rem"; // w-64
const COLLAPSED_W = "4rem"; // w-16

export default function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);
  const asideRef = useRef<HTMLElement>(null);

  // Adjust the main content's left offset without touching the layout file.
  // Finds the content element (#main-content, <main>, or the sidebar's next
  // sibling) and detects whether it uses margin-left or padding-left for the
  // sidebar gap. On collapse we override it inline; on expand we remove the
  // inline style so the original Tailwind class takes over again.
  useEffect(() => {
    const content =
      (document.getElementById("main-content") as HTMLElement | null) ??
      (document.querySelector("main") as HTMLElement | null) ??
      (asideRef.current?.nextElementSibling as HTMLElement | null);
    if (!content) return;

    const computed = window.getComputedStyle(content);
    const prop: "marginLeft" | "paddingLeft" =
      parseFloat(computed.marginLeft) >= 100 ? "marginLeft" : "paddingLeft";

    content.style.transition = "margin-left 200ms ease, padding-left 200ms ease";
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
        collapsed ? "w-16" : "w-64"
      )}
    >
      {/* Collapse toggle — sits on the sidebar's right border */}
      <button
        onClick={() => setCollapsed((c) => !c)}
        aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        className="absolute -right-3 top-16 z-50 w-6 h-6 rounded-full bg-zinc-900 border border-zinc-700 flex items-center justify-center text-zinc-400 hover:text-white hover:border-zinc-500 transition-colors shadow-md"
      >
        <ChevronLeftIcon
          className={clsx("w-3.5 h-3.5 transition-transform duration-200", collapsed && "rotate-180")}
        />
      </button>

      {/* Logo */}
      <div
        className={clsx(
          "flex items-center gap-3 py-5 border-b border-zinc-800",
          collapsed ? "px-3.5 justify-center" : "px-6"
        )}
      >
        <div className="w-9 h-9 rounded-md bg-white flex items-center justify-center overflow-hidden shrink-0">
          <Image
            src="/logo.png"
            alt="Market Lens Logo"
            width={30}
            height={30}
            className="object-contain"
          />
        </div>
        {!collapsed && (
          <div className="min-w-0">
            <h1 className="text-white font-bold text-sm tracking-tight leading-tight">Market Lens</h1>
            <p className="text-zinc-400 text-[10px] uppercase font-semibold tracking-wider leading-tight mt-0.5">
              Labour Intelligence
            </p>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navigation.map((item) => {
          const isActive =
            pathname === item.href ||
            (item.href !== "/" && pathname.startsWith(item.href));
          return (
            <Link
              key={item.name}
              href={item.href}
              title={collapsed ? item.name : undefined}
              className={clsx(
                "flex items-center gap-3 rounded-lg text-sm font-medium transition-all duration-150",
                collapsed ? "justify-center px-0 py-2.5" : "px-3 py-2.5",
                isActive
                  ? "bg-white text-black font-semibold shadow-sm"
                  : "text-zinc-400 hover:bg-zinc-900 hover:text-white"
              )}
            >
              <item.icon className="w-4 h-4 flex-shrink-0" />
              {!collapsed && item.name}
            </Link>
          );
        })}
      </nav>

      {/* Profile Footer */}
      <div className={clsx("py-4 border-t border-zinc-800 bg-zinc-950", collapsed ? "px-2" : "px-4")}>
        <div className={clsx("flex items-center gap-3", collapsed ? "justify-center" : "px-2")}>
          <div
            className="w-8 h-8 rounded-full bg-zinc-900 border border-zinc-700 flex items-center justify-center shrink-0"
            title={collapsed ? "Basuru Jagadakshi · Super Admin" : undefined}
          >
            <span className="text-zinc-200 text-xs font-bold">SK</span>
          </div>
          {!collapsed && (
            <div className="flex-1 min-w-0">
              <p className="text-zinc-100 text-xs font-semibold truncate">Basuru Jagadakshi</p>
              <p className="text-zinc-500 text-[10px] tracking-wide mt-0.5 truncate">Super Admin · TVEC</p>
            </div>
          )}
        </div>
      </div>
    </aside>
  );
}

// --- Inline Minimal SVG Icon components ---

function ChevronLeftIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="15 18 9 12 15 6" />
    </svg>
  );
}

function LayoutDashboardIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
  );
}

function BriefcaseIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="7" width="20" height="14" rx="2" />
      <path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2" />
    </svg>
  );
}

function BarChart3Icon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 3v18h18" /><path d="M18 17V9" /><path d="M13 17V5" /><path d="M8 17v-3" />
    </svg>
  );
}

function ScanSearchIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 7V5a2 2 0 0 1 2-2h2" />
      <path d="M17 3h2a2 2 0 0 1 2 2v2" />
      <path d="M21 17v2a2 2 0 0 1-2 2h-2" />
      <path d="M7 21H5a2 2 0 0 1-2-2v-2" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function GaugeIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="m12 14 4-4" />
      <path d="M3.34 19a10 10 0 1 1 17.32 0" />
    </svg>
  );
}

function ForecastingIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="3 17 8 12 12 15 16 8" />
      <line x1="16" y1="8" x2="21" y2="4" strokeDasharray="2 2" />
      <circle cx="21" cy="4" r="1.5" fill="currentColor" stroke="none" />
    </svg>
  );
}