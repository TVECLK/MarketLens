import type { Metadata } from "next";
import { ThunderIDProvider } from "@thunderid/nextjs/server";
import { Inter } from "next/font/google";
import "./globals.css";
import Sidebar from "@/components/layout/Sidebar";
import QueryProvider from "@/providers/query-provider";
import { scopeString } from "../lib/scopes";
import { SignedIn, SignedOut } from "@thunderid/nextjs";

export const dynamic = "force-dynamic";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "LMIS - Labour Market Intelligence System",
  description:
    "Gen-AI–enabled National Labour Market Demand & Vacancy Intelligence Platform – Sri Lanka",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
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
            <div className="flex min-h-screen">
              <SignedIn>
                <Sidebar />
                <main className="flex-1 ml-56 min-h-screen">{children}</main>
              </SignedIn>
              <SignedOut>
                <main className="flex-1 min-h-screen">{children}</main>
              </SignedOut>
            </div>
          </QueryProvider>
        </ThunderIDProvider>
      </body>
    </html>
  );
}
