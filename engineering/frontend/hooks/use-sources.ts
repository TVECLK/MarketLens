"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type {
  CreateSourcePayload,
  Source,
  SourceFilters,
  SourceListResponse,
} from "@/types/source";
import { ENV } from "@/config/env";
import { fetchJson } from "@/lib/fetch-json";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;
const SOURCES_URL = `${BASE}/sources`;

/** Root of every source query key, e.g. `["sources", "list", filters]`. */
export const SOURCES_QUERY_KEY = "sources";

/** The query key of a source list. */
export function sourcesListKey(filters: SourceFilters = {}) {
  return [SOURCES_QUERY_KEY, "list", filters] as const;
}

/** `/sources` with the backend's `?type=` and `?is-active=` filters. */
export function sourcesListUrl(filters: SourceFilters = {}): string {
  const params = new URLSearchParams();
  if (filters.type) params.set("type", filters.type);
  if (filters.isActive !== undefined) params.set("is-active", String(filters.isActive));
  const query = params.toString();
  return query ? `${SOURCES_URL}?${query}` : SOURCES_URL;
}

/** Lists sources (deleted ones are never returned). */
export function useSources(filters: SourceFilters = {}) {
  return useQuery<SourceListResponse>({
    queryKey: sourcesListKey(filters),
    queryFn: ({ signal }) =>
      fetchJson<SourceListResponse>(sourcesListUrl(filters), {
        signal,
        errorMessage: "Failed to fetch sources",
      }),
  });
}

/** Adds a source. Re-using a deleted source's name restores it (same id, new settings). */
export function useCreateSource() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateSourcePayload) =>
      fetchJson<Source>(SOURCES_URL, {
        method: "POST",
        body: payload,
        errorMessage: "Failed to create source",
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [SOURCES_QUERY_KEY] }),
  });
}

/** Renames a manual source. The backend rejects renaming an auto source. */
export function useRenameSource() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, name }: { id: number; name: string }) =>
      fetchJson<Source>(`${SOURCES_URL}/${id}`, {
        method: "PUT",
        body: { name },
        errorMessage: "Failed to rename source",
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [SOURCES_QUERY_KEY] }),
  });
}


export function useUpdateSourceStatus() {
  const queryClient = useQueryClient();
  const listsKey = [SOURCES_QUERY_KEY, "list"];

  /** Sets one source's `is_active` in every cached list, leaving the other sources untouched. */
  const setCachedStatus = (id: number, isActive: boolean) =>
    queryClient.setQueriesData<SourceListResponse>({ queryKey: listsKey }, (data) =>
      data && {
        ...data,
        sources: data.sources.map((s) => (s.id === id ? { ...s, is_active: isActive } : s)),
      },
    );

  return useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      fetchJson<Source>(`${SOURCES_URL}/${id}/status`, {
        method: "PATCH",
        body: { is_active: isActive },
        errorMessage: "Failed to update source status",
      }),
    onMutate: async ({ id, isActive }): Promise<{ previous: boolean | undefined }> => {
      await queryClient.cancelQueries({ queryKey: listsKey });
      // This source's status before the change, from whichever cached list has it.
      const previous = queryClient
        .getQueriesData<SourceListResponse>({ queryKey: listsKey })
        .flatMap(([, data]) => data?.sources ?? [])
        .find((s) => s.id === id)?.is_active;
      setCachedStatus(id, isActive);
      return { previous };
    },
    onError: (_err, { id }, context) => {
      if (context?.previous !== undefined) setCachedStatus(id, context.previous);
    },
    // Refetch either way: a filtered list (e.g. active only) may now include or exclude the source.
    onSettled: () => queryClient.invalidateQueries({ queryKey: [SOURCES_QUERY_KEY] }),
  });
}

/** Soft-deletes a source. Its jobs stay linked to it; adding the same name again restores it. */
export function useDeleteSource() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      fetchJson<{ message: string }>(`${SOURCES_URL}/${id}`, {
        method: "DELETE",
        errorMessage: "Failed to delete source",
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [SOURCES_QUERY_KEY] }),
  });
}
