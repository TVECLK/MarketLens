import {
  IndustryListResponse,
  IndustryDivisionChildrenResponse,
  IndustryGroupChildrenResponse,
  IndustryClassChildrenResponse,
  IndustrySubclassChildrenResponse,
} from "@/types/industry";
import { INDUSTRY_QUERY_KEYS } from "@/lib/kpi-config";
import { useHierarchyRoot, useHierarchyChildren } from "@/hooks/use-hierarchy-queries";

const [INDUSTRY_SECTORS, INDUSTRY_DIVISIONS, INDUSTRY_GROUPS, INDUSTRY_CLASSES, INDUSTRY_SUBCLASSES] = INDUSTRY_QUERY_KEYS;

export function useIndustrySectors(fromDate: string, toDate: string) {
  return useHierarchyRoot<IndustryListResponse>("industry", INDUSTRY_SECTORS, fromDate, toDate);
}

export function useIndustryDivisions(industrySectorId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<IndustryDivisionChildrenResponse>("industry", INDUSTRY_SECTORS, INDUSTRY_DIVISIONS, industrySectorId, fromDate, toDate);
}

export function useIndustryGroups(industryDivisionId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<IndustryGroupChildrenResponse>("industry", INDUSTRY_DIVISIONS, INDUSTRY_GROUPS, industryDivisionId, fromDate, toDate);
}

export function useIndustryClasses(industryGroupId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<IndustryClassChildrenResponse>("industry", INDUSTRY_GROUPS, INDUSTRY_CLASSES, industryGroupId, fromDate, toDate);
}

export function useIndustrySubclasses(industryClassId: number | null, fromDate: string, toDate: string) {
  return useHierarchyChildren<IndustrySubclassChildrenResponse>("industry", INDUSTRY_CLASSES, INDUSTRY_SUBCLASSES, industryClassId, fromDate, toDate);
}
