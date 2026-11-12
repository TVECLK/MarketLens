"use client";
import { useThunderID } from "@thunderid/nextjs";
import Sidebar from "./Sidebar";
import SessionWatcher from "./SessionWatcher";
import type { PageAccess } from "@/lib/scopes";

export default function AppShell({
  pageAccess,
  children,
}: {
  /**
   * The page scopes the user has, checked on the server (root layout), so a user with no admin page
   * never sees the sidebar, and a link the user can't open never shows, not even briefly.
   */
  pageAccess: PageAccess;
  children: React.ReactNode;
}) {
  const { isSignedIn } = useThunderID();
  // Shown with any page scope. Also require `isSignedIn`, so the sidebar disappears as soon as the
  // user signs out on the client.
  const showSidebar = !!isSignedIn && Object.values(pageAccess).some(Boolean);

  return (
    <div className="flex min-h-screen">
      {/* Notices a sign-out done in another tab when this one is used again. */}
      {isSignedIn && <SessionWatcher />}
      {showSidebar && <Sidebar pageAccess={pageAccess} />}
      <main
        id="main-content"
        className={showSidebar ? "flex-1 min-h-screen ml-56" : "flex-1 min-h-screen"}
      >
        {children}
      </main>
    </div>
  );
}
