import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
import { STALE_TIME } from "@/config/query";
import {
  MajorGroupListResponse,
  SubMajorGroupChildrenResponse,
  MinorGroupChildrenResponse,
  UnitGroupChildrenResponse,
  OccupationGroupChildrenResponse,
  OccupationAnalysisResponse,
  OccupationSkillsResponse
} from "@/types/occupation";
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

async function fetchOccupationSkills(
  level: string,
  id: number,
  fromDate: string,
  toDate: string,
): Promise<OccupationSkillsResponse> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate, limit: "50", offset: "0" });
  const res = await fetch(`${BASE}/occupation/skills/${level}/${id}?${params.toString()}`);
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(getApiErrorMessage(body, "Failed to fetch occupation skills"));
  }
  return res.json();
}

export function useOccupationSkills(
  level: string | null,
  id: number | null,
  fromDate: string,
  toDate: string,
  enabled: boolean,
) {
  return useQuery<OccupationSkillsResponse>({
    queryKey: ["occupation-skills", level, id, fromDate, toDate],
    queryFn: () => fetchOccupationSkills(level as string, id as number, fromDate, toDate),
    enabled: enabled && level !== null && id !== null && Boolean(fromDate) && Boolean(toDate) && fromDate <= toDate,
    staleTime: STALE_TIME,
  });
}

async function fetchOccupationAnalysis(
  level: string,
  id: number,
  fromDate: string,
  toDate: string,
): Promise<OccupationAnalysisResponse> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
  const res = await fetch(`${BASE}/analysis/occupation/${level}/${id}?${params.toString()}`);
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(getApiErrorMessage(body, "Failed to fetch occupation analysis"));
  }
  return res.json();
}

export function useOccupationAnalysis(
  level: string | null,
  id: number | null,
  fromDate: string,
  toDate: string,
) {
  return useQuery<OccupationAnalysisResponse>({
    queryKey: ["occupation-analysis", level, id, fromDate, toDate],
    queryFn: () => fetchOccupationAnalysis(level as string, id as number, fromDate, toDate),
    enabled: level !== null && id !== null && Boolean(fromDate) && Boolean(toDate) && fromDate <= toDate,
    staleTime: STALE_TIME,
  });
}

export function useMajorGroup(fromDate: string, toDate: string) {
  return useQuery<MajorGroupListResponse>({
    queryKey: ["occupation", "major-groups", fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/major-groups?${params.toString()}`);
    },
    enabled: Boolean(fromDate) && Boolean(toDate),
    staleTime: STALE_TIME,
  });
}

export function useSubMajorGroups(majorGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<SubMajorGroupChildrenResponse>({
    queryKey: ["occupation", "sub-major-groups", majorGroupId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/major-groups/${majorGroupId}/sub-major-groups?${params.toString()}`);
    },
    enabled: majorGroupId !== null && Boolean(fromDate) && Boolean(toDate),
    staleTime: STALE_TIME,
  });
}

export function useMinorGroups(subMajorGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<MinorGroupChildrenResponse>({
    queryKey: ["occupation", "minor-groups", subMajorGroupId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/sub-major-groups/${subMajorGroupId}/minor-groups?${params.toString()}`);
    },
    enabled: subMajorGroupId !== null && Boolean(fromDate) && Boolean(toDate),
    staleTime: STALE_TIME,
  });
}

export function useUnitGroups(minorGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<UnitGroupChildrenResponse>({
    queryKey: ["occupation", "unit-groups", minorGroupId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/minor-groups/${minorGroupId}/unit-groups?${params.toString()}`);
    },
    enabled: minorGroupId !== null && Boolean(fromDate) && Boolean(toDate),
    staleTime: STALE_TIME,
  });
}

export function useOccupationGroups(unitGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<OccupationGroupChildrenResponse>({
    queryKey: ["occupation", "occupation-groups", unitGroupId, fromDate, toDate],
    queryFn: () => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/unit-groups/${unitGroupId}/occupation-groups?${params.toString()}`);
    },
    enabled: unitGroupId !== null && Boolean(fromDate) && Boolean(toDate),
    staleTime: STALE_TIME,
  });
}