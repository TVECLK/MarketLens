import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
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

async function fetchJson<T>(url: string, signal?: AbortSignal): Promise<T> {
  const res = await fetch(url, { signal });
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
  signal?: AbortSignal,
): Promise<OccupationSkillsResponse> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate, limit: "50", offset: "0" });
  const res = await fetch(`${BASE}/occupation/skills/${level}/${id}?${params.toString()}`, { signal });
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
    queryFn: ({ signal }) => fetchOccupationSkills(level as string, id as number, fromDate, toDate, signal),
    enabled: enabled && level !== null && id !== null && Boolean(fromDate) && Boolean(toDate) && fromDate <= toDate,
  });
}

async function fetchOccupationAnalysis(
  level: string,
  id: number,
  fromDate: string,
  toDate: string,
  signal?: AbortSignal,
): Promise<OccupationAnalysisResponse> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
  const res = await fetch(`${BASE}/analysis/occupation/${level}/${id}?${params.toString()}`, { signal });
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
    queryFn: ({ signal }) => fetchOccupationAnalysis(level as string, id as number, fromDate, toDate, signal),
    enabled: level !== null && id !== null && Boolean(fromDate) && Boolean(toDate) && fromDate <= toDate,
  });
}

export function useMajorGroup(fromDate: string, toDate: string) {
  return useQuery<MajorGroupListResponse>({
    queryKey: ["occupation", "major-groups", fromDate, toDate],
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/major-groups?${params.toString()}`, signal);
    },
    enabled: Boolean(fromDate) && Boolean(toDate),
  });
}

export function useSubMajorGroups(majorGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<SubMajorGroupChildrenResponse>({
    queryKey: ["occupation", "sub-major-groups", majorGroupId, fromDate, toDate],
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/major-groups/${majorGroupId}/sub-major-groups?${params.toString()}`, signal);
    },
    enabled: majorGroupId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}

export function useMinorGroups(subMajorGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<MinorGroupChildrenResponse>({
    queryKey: ["occupation", "minor-groups", subMajorGroupId, fromDate, toDate],
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/sub-major-groups/${subMajorGroupId}/minor-groups?${params.toString()}`, signal);
    },
    enabled: subMajorGroupId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}

export function useUnitGroups(minorGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<UnitGroupChildrenResponse>({
    queryKey: ["occupation", "unit-groups", minorGroupId, fromDate, toDate],
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/minor-groups/${minorGroupId}/unit-groups?${params.toString()}`, signal);
    },
    enabled: minorGroupId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}

export function useOccupationGroups(unitGroupId: number | null, fromDate: string, toDate: string) {
  return useQuery<OccupationGroupChildrenResponse>({
    queryKey: ["occupation", "occupation-groups", unitGroupId, fromDate, toDate],
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate });
      return fetchJson(`${BASE}/occupation/unit-groups/${unitGroupId}/occupation-groups?${params.toString()}`, signal);
    },
    enabled: unitGroupId !== null && Boolean(fromDate) && Boolean(toDate),
  });
}