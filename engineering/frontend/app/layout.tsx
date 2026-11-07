import type { Metadata } from "next";
import { ThunderIDProvider } from "@thunderid/nextjs/server";
import { Inter } from "next/font/google";
import "./globals.css";
import QueryProvider from "@/providers/query-provider";
import AppShell from "@/components/layout/AppShell";
import { PAGE_SCOPES, scopeString } from "../lib/scopes";
import { hasServerPermission } from "@/lib/server-permissions";

export const dynamic = "force-dynamic";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "LMIS - Labour Market Intelligence System",
  description:
    "Gen-AI–enabled National Labour Market Demand & Vacancy Intelligence Platform – Sri Lanka",
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // Read on the server (AppShell is a client component and can't read the session cookie).
  const canAccessKpis = await hasServerPermission(PAGE_SCOPES.kpis);

  return (
    <html lang="en">
      <body className={inter.className}>
        <ThunderIDProvider
          clientId={process.env.NEXT_PUBLIC_THUNDERID_CLIENT_ID}
          baseUrl={process.env.NEXT_PUBLIC_THUNDERID_BASE_URL}
          scopes={scopeString}
        >
          {/* Wrap your entire application structure with the query cache context provider */}
          <QueryProvider>
            <AppShell canAccessKpis={canAccessKpis}>{children}</AppShell>
          </QueryProvider>
        </ThunderIDProvider>
      </body>
    </html>
  );
}
