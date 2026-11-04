import { useQuery } from "@tanstack/react-query";
import { ENV } from "@/config/env";
import {
  MajorGroupListResponse,
  SubMajorGroupChildrenResponse,
  MinorGroupChildrenResponse,
  UnitGroupChildrenResponse,
  OccupationGroupChildrenResponse,
  OccupationSkillsResponse
} from "@/types/occupation";
import { fetchJson } from "@/lib/fetch-json";
import { OCCUPATION_QUERY_KEYS } from "@/lib/kpi-config";
import { useHierarchyRoot, useHierarchyChildren } from "@/hooks/use-hierarchy-queries";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;

const [MAJOR_GROUPS, SUB_MAJOR_GROUPS, MINOR_GROUPS, UNIT_GROUPS, OCCUPATION_GROUPS] = OCCUPATION_QUERY_KEYS;

async function fetchOccupationSkills(
  level: string,
  id: number,
  fromDate: string,
  toDate: string,
  signal?: AbortSignal,
): Promise<OccupationSkillsResponse> {
  const params = new URLSearchParams({ "from-date": fromDate, "to-date": toDate, limit: "50", offset: "0" });
  return fetchJson<OccupationSkillsResponse>(`${BASE}/occupation/skills/${level}/${id}?${params.toString()}`, { signal, errorMessage: "Failed to fetch occupation skills" });
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

export function useMajorGroup(fromDate: string, toDate: string) {
  return useHierarchyRoot<MajorGroupListResponse>("occupation", MAJOR_GROUPS, fromDate, toDate);
}

export function useSubMajorGroups(majorGroupId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<SubMajorGroupChildrenResponse>("occupation", MAJOR_GROUPS, SUB_MAJOR_GROUPS, majorGroupId, fromDate, toDate);
}

export function useMinorGroups(subMajorGroupId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<MinorGroupChildrenResponse>("occupation", SUB_MAJOR_GROUPS, MINOR_GROUPS, subMajorGroupId, fromDate, toDate);
}

export function useUnitGroups(minorGroupId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<UnitGroupChildrenResponse>("occupation", MINOR_GROUPS, UNIT_GROUPS, minorGroupId, fromDate, toDate);
}

export function useOccupationGroups(unitGroupId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<OccupationGroupChildrenResponse>("occupation", UNIT_GROUPS, OCCUPATION_GROUPS, unitGroupId, fromDate, toDate);
}
