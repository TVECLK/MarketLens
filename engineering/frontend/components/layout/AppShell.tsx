"use client";
import { useThunderID } from "@thunderid/nextjs";
import Sidebar from "./Sidebar";

export default function AppShell({ children }: { children: React.ReactNode }) {
  const { isSignedIn } = useThunderID();
  return (
    <div className="flex min-h-screen">
      {isSignedIn && <Sidebar />}
      <main
        id="main-content"
        className={isSignedIn ? "flex-1 min-h-screen ml-64" : "flex-1 min-h-screen"}
      >
        {children}
      </main>
    </div>
  );
}
