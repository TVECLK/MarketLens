"use client";

import {
  useMajorGroups, useCreateMajorGroup, useUpdateMajorGroup, useDeleteMajorGroup,
  useCreateSubMajorGroup, useUpdateSubMajorGroup, useDeleteSubMajorGroup,
  useCreateMinorGroup, useUpdateMinorGroup, useDeleteMinorGroup,
  useCreateUnitGroup, useUpdateUnitGroup, useDeleteUnitGroup,
  useCreateOccupationGroup, useUpdateOccupationGroup, useDeleteOccupationGroup,

  useIndustrySectors, useCreateIndustrySector, useUpdateIndustrySector, useDeleteIndustrySector,
  useCreateIndustryDivision, useUpdateIndustryDivision, useDeleteIndustryDivision,
  useCreateIndustryGroup, useUpdateIndustryGroup, useDeleteIndustryGroup,
  useCreateIndustryClass, useUpdateIndustryClass, useDeleteIndustryClass,
  useCreateIndustrySubclass, useUpdateIndustrySubclass, useDeleteIndustrySubclass,
} from "@/hooks/use-kpi";
import { useActionPermissions } from "@/hooks/use-action-permissions";
import type { HierarchyId } from "@/lib/kpi-config";
import { normTop } from "@/lib/kpi-normalizers";
import { RESOURCE_SCOPES } from "@/lib/scopes";
import type { HierarchyLevelMutations } from "@/types/kpi-admin";
import HierarchyTreeView from "@/components/kpis/HierarchyTreeView";

// Only the top level is loaded as a list; every other level is fetched by
// parent when its parent is expanded (see use-admin-hierarchy-children.ts).

function OccupationRoute() {
  const majorGroupsQ = useMajorGroups();

  const levels: HierarchyLevelMutations[] = [
    {
      createMut: useCreateMajorGroup(), updateMut: useUpdateMajorGroup(), deleteMut: useDeleteMajorGroup(),
      permissions: useActionPermissions(RESOURCE_SCOPES.majorGroups), parentField: null,
    },
    {
      createMut: useCreateSubMajorGroup(), updateMut: useUpdateSubMajorGroup(), deleteMut: useDeleteSubMajorGroup(),
      permissions: useActionPermissions(RESOURCE_SCOPES.subMajorGroups), parentField: "major_group_id",
    },
    {
      createMut: useCreateMinorGroup(), updateMut: useUpdateMinorGroup(), deleteMut: useDeleteMinorGroup(),
      permissions: useActionPermissions(RESOURCE_SCOPES.minorGroups), parentField: "sub_major_group_id",
    },
    {
      createMut: useCreateUnitGroup(), updateMut: useUpdateUnitGroup(), deleteMut: useDeleteUnitGroup(),
      permissions: useActionPermissions(RESOURCE_SCOPES.unitGroups), parentField: "minor_group_id",
    },
    {
      createMut: useCreateOccupationGroup(), updateMut: useUpdateOccupationGroup(), deleteMut: useDeleteOccupationGroup(),
      permissions: useActionPermissions(RESOURCE_SCOPES.occupationGroups), parentField: "unit_group_id",
    },
  ];

  return (
    <HierarchyTreeView
      hierarchyId="occupation"
      roots={{
        isLoading: majorGroupsQ.isLoading,
        isError: majorGroupsQ.isError,
        items: normTop(majorGroupsQ.data?.major_groups),
      }}
      levels={levels}
    />
  );
}

function IndustryRoute() {
  const industrySectorsQ = useIndustrySectors();

  const levels: HierarchyLevelMutations[] = [
    {
      createMut: useCreateIndustrySector(), updateMut: useUpdateIndustrySector(), deleteMut: useDeleteIndustrySector(),
      permissions: useActionPermissions(RESOURCE_SCOPES.industrySectors), parentField: null,
    },
    {
      createMut: useCreateIndustryDivision(), updateMut: useUpdateIndustryDivision(), deleteMut: useDeleteIndustryDivision(),
      permissions: useActionPermissions(RESOURCE_SCOPES.industryDivisions), parentField: "industry_sector_id",
    },
    {
      createMut: useCreateIndustryGroup(), updateMut: useUpdateIndustryGroup(), deleteMut: useDeleteIndustryGroup(),
      permissions: useActionPermissions(RESOURCE_SCOPES.industryGroups), parentField: "industry_division_id",
    },
    {
      createMut: useCreateIndustryClass(), updateMut: useUpdateIndustryClass(), deleteMut: useDeleteIndustryClass(),
      permissions: useActionPermissions(RESOURCE_SCOPES.industryClasses), parentField: "industry_group_id",
    },
    {
      createMut: useCreateIndustrySubclass(), updateMut: useUpdateIndustrySubclass(), deleteMut: useDeleteIndustrySubclass(),
      permissions: useActionPermissions(RESOURCE_SCOPES.industrySubClasses), parentField: "industry_class_id",
    },
  ];

  return (
    <HierarchyTreeView
      hierarchyId="industry"
      roots={{
        isLoading: industrySectorsQ.isLoading,
        isError: industrySectorsQ.isError,
        items: normTop(industrySectorsQ.data?.industry_sectors),
      }}
      levels={levels}
    />
  );
}

export default function HierarchyRoute({ hierarchyId }: { hierarchyId: HierarchyId }) {
  return hierarchyId === "occupation" ? <OccupationRoute /> : <IndustryRoute />;
}
