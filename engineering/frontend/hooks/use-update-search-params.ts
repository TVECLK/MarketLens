"use client";

import { useCallback, useEffect, useRef } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

export type SearchParamUpdates = Record<string, string | null>;


export function useUpdateSearchParams(): (updates: SearchParamUpdates) => void {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const latestQuery = useRef(searchParams.toString());

  useEffect(() => {
    latestQuery.current = searchParams.toString();
  }, [searchParams]);

  return useCallback(
    (updates) => {
      const params = new URLSearchParams(latestQuery.current);
      for (const [key, value] of Object.entries(updates)) {
        if (value === null) params.delete(key);
        else params.set(key, value);
      }
      const query = params.toString();
      latestQuery.current = query;
      router.replace(query ? `${pathname}?${query}` : pathname, { scroll: false });
    },
    [router, pathname],
  );
}
