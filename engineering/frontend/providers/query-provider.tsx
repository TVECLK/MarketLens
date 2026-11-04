"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, ReactNode } from "react";
import { STALE_TIME } from "@/config/query";

export default function QueryProvider({ children }: { children: ReactNode }) {
  // Using useState ensures Next.js only instantiates the cache once per session
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: STALE_TIME, // cache freshness default
            refetchOnWindowFocus: false, // Prevents aggressive network calls on tab switching
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}