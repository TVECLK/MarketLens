import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
import {
  DashboardOverview,
} from "@/types/dashboard";
import { isValidDateRange } from "@/lib/date-range";
import { fetchJson } from "@/lib/fetch-json";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;

async function fetchDashboardOverview(fromDate: string, toDate: string, signal?: AbortSignal): Promise<DashboardOverview> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
  return fetchJson<DashboardOverview>(`${BASE}/dashboard/overview?${params.toString()}`, { signal, errorMessage: "Failed to fetch dashboard overview" });
}

export function useDashboardOverview(fromDate: string, toDate: string) {
  return useQuery<DashboardOverview>({
    queryKey: ["dashboard", "overview", fromDate, toDate],
    queryFn: ({ signal }) => fetchDashboardOverview(fromDate, toDate, signal),
    enabled: isValidDateRange(fromDate, toDate),
  });
}