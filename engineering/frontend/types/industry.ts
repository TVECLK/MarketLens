import type { AnalysisResponse, HierarchyNode } from "./analysis";

export type { HierarchyNode } from "./analysis";

export interface IndustryDivisionChildrenResponse {
  industry_sector_id: number;
  count: number;
  industry_divisions: HierarchyNode[];
}
export interface IndustryGroupChildrenResponse {
  industry_division_id: number;
  count: number;
  industry_groups: HierarchyNode[];
}
export interface IndustryClassChildrenResponse {
  industry_group_id: number;
  count: number;
  industry_classes: HierarchyNode[];
}
export interface IndustrySubclassChildrenResponse {
  industry_class_id: number;
  count: number;
  industry_subclasses: HierarchyNode[];
}

export type IndustryAnalysisResponse = AnalysisResponse;

export interface Industry {
  id:   number;
  name: string;
}

export interface IndustryListResponse {
  count:      number;
  industry_sectors: Industry[];
}
