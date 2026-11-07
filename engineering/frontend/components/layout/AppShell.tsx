"use client";
import { useThunderID } from "@thunderid/nextjs";
import Sidebar from "./Sidebar";

export default function AppShell({
  canAccessKpis,
  children,
}: {
  /** Checked on the server (root layout), so a user without KPI access never sees the sidebar, not even briefly. */
  canAccessKpis: boolean;
  children: React.ReactNode;
}) {
  const { isSignedIn } = useThunderID();
  // Also require `isSignedIn`, so the sidebar disappears as soon as the user signs out on the client.
  const showSidebar = !!isSignedIn && canAccessKpis;

  return (
    <div className="flex min-h-screen">
      {showSidebar && <Sidebar />}
      <main
        id="main-content"
        className={showSidebar ? "flex-1 min-h-screen ml-56" : "flex-1 min-h-screen"}
      >
        {children}
      </main>
    </div>
  );
}
