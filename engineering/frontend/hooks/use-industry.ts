import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
import { IndustryListResponse, 
  IndustryDivisionChildrenResponse,
  IndustryGroupChildrenResponse,
  IndustryClassChildrenResponse,
  IndustrySubclassChildrenResponse,
  IndustryAnalysisResponse,
} from "@/types/industry";
import { getApiErrorMessage } from "@/lib/errors";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;

async function fetchJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(getApiErrorMessage(body, `Request failed: ${url}`));
  }
  return res.json();
}

export function useIndustrySectors(fromDate: string, toDate: string) {
  return useQuery<IndustryListResponse>({
    queryKey: ["industry", "sectors", fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/industry/industry-sectors?${params.toString()}`);
    },
    enabled: Boolean(fromDate) && Boolean(toDate),
  });
}

export function useIndustryDivisions(industrySectorId: number | null, fromDate: string, toDate: string) {
  return useQuery<IndustryDivisionChildrenResponse>({
    queryKey: ["industry", "industry-divisions", industrySectorId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/industry/industry-sectors/${industrySectorId}/industry-divisions?${params.toString()}`);
    },
    enabled: industrySectorId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}

export function useIndustryGroups(industryDivisionId: number | null, fromDate: string, toDate: string) {
  return useQuery<IndustryGroupChildrenResponse>({
    queryKey: ["industry", "industry-groups", industryDivisionId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/industry/industry-divisions/${industryDivisionId}/industry-groups?${params.toString()}`);
    },
    enabled: industryDivisionId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}

export function useIndustryClasses(industryGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<IndustryClassChildrenResponse>({
    queryKey: ["industry", "industry-classes", industryGroupId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/industry/industry-groups/${industryGroupId}/industry-classes?${params.toString()}`);
    },
    enabled: industryGroupId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}

export function useIndustrySubclasses(industryClassId: number | null, fromDate: string, toDate: string) {
  return useQuery<IndustrySubclassChildrenResponse>({
    queryKey: ["industry", "industry-subclasses", industryClassId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/industry/industry-classes/${industryClassId}/industry-subclasses?${params.toString()}`);
    },
    enabled: industryClassId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}

async function fetchIndustryAnalysis(
  level: string,
  id: number,
  fromDate: string,
  toDate: string,
): Promise<IndustryAnalysisResponse> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
  const res = await fetch(`${BASE}/analysis/industry/${level}/${id}?${params.toString()}`);
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(getApiErrorMessage(body, "Failed to fetch industry analysis"));
  }
  return res.json();
}

export function useIndustryAnalysis(
  level: string | null,
  id: number | null,
  fromDate: string,
  toDate: string,
) {
  return useQuery<IndustryAnalysisResponse>({
    queryKey: ["industry-analysis", level, id, fromDate, toDate],
    queryFn: () => fetchIndustryAnalysis(level as string, id as number, fromDate, toDate),
    enabled: level !== null && id !== null && Boolean(fromDate) && Boolean(toDate) && fromDate <= toDate,
  });
}