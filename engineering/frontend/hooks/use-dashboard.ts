import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
import {
  DashboardOverview,
} from "@/types/dashboard";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;

async function fetchDashboardOverview(fromDate: string, toDate: string): Promise<DashboardOverview> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
  const res = await fetch(`${BASE}/dashboard/overview?${params.toString()}`);
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error || body?.details || "Failed to fetch dashboard overview");
  }
  return res.json();
}

export function useDashboardOverview(fromDate: string, toDate: string) {
  return useQuery<DashboardOverview>({
    queryKey: ["dashboard", "overview", fromDate, toDate],
    queryFn: () => fetchDashboardOverview(fromDate, toDate),
    enabled: Boolean(fromDate) && Boolean(toDate) && fromDate <= toDate,
    staleTime: 1000 * 60 * 5,
  });
}