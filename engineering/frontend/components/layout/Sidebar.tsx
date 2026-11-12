"use client";

import { SignedIn } from "@thunderid/nextjs";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { clsx } from "clsx";
import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import { ChevronLeftIcon } from "@heroicons/react/16/solid";
import { ChartBarIcon, CircleStackIcon, Squares2X2Icon } from "@heroicons/react/24/outline";
import { THEME } from "@/lib/theme";
import type { PageAccess, PageScopeKey } from "@/lib/scopes";

interface NavItem {
  name: string;
  href: string;
  icon: typeof Squares2X2Icon;
  /** Only shown to signed-in users. */
  protected?: boolean;
  /** Only shown when the user has this page scope. */
  page?: PageScopeKey;
}

const navigation: NavItem[] = [
  { name: "Dashboard", href: "/", icon: Squares2X2Icon },
  { name: "KPIs", href: "/kpis", icon: ChartBarIcon, protected: true, page: "kpis" },
  { name: "Data Sources", href: "/data-sources", icon: CircleStackIcon, protected: true, page: "datasource" },
];

const COLLAPSED_W = "4rem"; // w-16

export default function Sidebar({ pageAccess }: { pageAccess: PageAccess }) {
  const visibleNavigation = navigation.filter((item) => !item.page || pageAccess[item.page]);
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
        "fixed inset-y-0 left-0 z-50 border-r flex flex-col font-sans transition-all duration-200",
        THEME.sidebar.bg,
        THEME.sidebar.border,
        collapsed ? "w-16" : "w-56",
      )}
    >
      {/* Collapse toggle — sits on the sidebar's right border */}
      <button
        onClick={() => setCollapsed((c) => !c)}
        aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        className={clsx(
          "absolute -right-3 top-16 z-50 w-6 h-6 rounded-full border flex items-center justify-center transition-colors shadow-md cursor-pointer",
          THEME.sidebar.toggle,
        )}
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
          "flex items-center gap-3 py-5 border-b",
          THEME.sidebar.border,
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
            <p className={clsx("text-[9px] uppercase font-semibold tracking-wider leading-tight mt-0.5", THEME.sidebar.subtext)}>
              Labour Intelligence
            </p>
          </div>
        )}
      </div>

      {/* Navigation Links */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {visibleNavigation.map((item) => {
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
                isActive ? THEME.sidebar.linkActive : THEME.sidebar.link,
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
