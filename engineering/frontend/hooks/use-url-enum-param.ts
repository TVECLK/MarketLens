"use client";

import { useCallback } from "react";
import { useSearchParams } from "next/navigation";
import { parseEnumParam } from "@/lib/url-params";
import { useUpdateSearchParams } from "@/hooks/use-update-search-params";

/**
 * A single-choice UI state (view mode, active tab, ...) stored in the URL as
 * `?<key>=<value>`. Missing or unrecognised values fall back to `fallback`.
 */
export function useUrlEnumParam<T extends string>(
  key: string,
  allowed: readonly T[],
  fallback: T,
): [T, (value: T) => void] {
  const searchParams = useSearchParams();
  const updateSearchParams = useUpdateSearchParams();
  const value = parseEnumParam(searchParams.get(key), allowed, fallback);

  const setValue = useCallback(
    (next: T) => updateSearchParams({ [key]: next }),
    [key, updateSearchParams],
  );

  return [value, setValue];
}
