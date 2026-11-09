"use client";

import {
  useMajorGroups, useSubMajorGroups, useMinorGroups, useUnitGroups, useCreateMajorGroup, useUpdateMajorGroup, useDeleteMajorGroup,
  useCreateSubMajorGroup, useUpdateSubMajorGroup, useDeleteSubMajorGroup,
  useCreateMinorGroup, useUpdateMinorGroup, useDeleteMinorGroup,
  useCreateUnitGroup, useUpdateUnitGroup, useDeleteUnitGroup,
  useCreateOccupationGroup, useUpdateOccupationGroup, useDeleteOccupationGroup,

  useIndustrySectors, useIndustryDivisions, useIndustryGroups, useIndustryClasses, useCreateIndustrySector, useUpdateIndustrySector, useDeleteIndustrySector,
  useCreateIndustryDivision, useUpdateIndustryDivision, useDeleteIndustryDivision,
  useCreateIndustryGroup, useUpdateIndustryGroup, useDeleteIndustryGroup,
  useCreateIndustryClass, useUpdateIndustryClass, useDeleteIndustryClass,
  useCreateIndustrySubclass, useUpdateIndustrySubclass, useDeleteIndustrySubclass,
} from "@/hooks/use-kpi";
import { useActionPermissions } from "@/hooks/use-action-permissions";
import type { HierarchyId } from "@/lib/kpi-config";
import { normChild, normTop } from "@/lib/kpi-normalizers";
import { RESOURCE_SCOPES } from "@/lib/scopes";
import type { HierarchyLevelMutations } from "@/types/kpi-admin";
import HierarchyTreeView from "@/components/kpis/HierarchyTreeView";

// The tree fetches each level by parent when its parent is expanded (see
// use-admin-hierarchy-children.ts). Every level that can be someone's parent is
// also loaded in full, to fill the edit form's "Parent" dropdown; the leaf level
// never is, so it isn't (it's the only paginated one).

function OccupationRoute() {
  const majorGroupsQ = useMajorGroups();
  const subMajorGroupsQ = useSubMajorGroups();
  const minorGroupsQ = useMinorGroups();
  const unitGroupsQ = useUnitGroups();

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
      lists={[
        { isLoading: majorGroupsQ.isLoading, isError: majorGroupsQ.isError, items: normTop(majorGroupsQ.data?.major_groups) },
        {
          isLoading: subMajorGroupsQ.isLoading, isError: subMajorGroupsQ.isError,
          items: normChild(subMajorGroupsQ.data?.sub_major_groups, "major_group_id", "major_group"),
        },
        {
          isLoading: minorGroupsQ.isLoading, isError: minorGroupsQ.isError,
          items: normChild(minorGroupsQ.data?.minor_groups, "sub_major_group_id", "sub_major_group"),
        },
        {
          isLoading: unitGroupsQ.isLoading, isError: unitGroupsQ.isError,
          items: normChild(unitGroupsQ.data?.unit_groups, "minor_group_id", "minor_group"),
        },
      ]}
      levels={levels}
    />
  );
}

function IndustryRoute() {
  const industrySectorsQ = useIndustrySectors();
  const industryDivisionsQ = useIndustryDivisions();
  const industryGroupsQ = useIndustryGroups();
  const industryClassesQ = useIndustryClasses();

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
      lists={[
        { isLoading: industrySectorsQ.isLoading, isError: industrySectorsQ.isError, items: normTop(industrySectorsQ.data?.industry_sectors) },
        {
          isLoading: industryDivisionsQ.isLoading, isError: industryDivisionsQ.isError,
          items: normChild(industryDivisionsQ.data?.industry_divisions, "industry_sector_id", "industry_sector"),
        },
        {
          isLoading: industryGroupsQ.isLoading, isError: industryGroupsQ.isError,
          items: normChild(industryGroupsQ.data?.industry_groups, "industry_division_id", "industry_division"),
        },
        {
          isLoading: industryClassesQ.isLoading, isError: industryClassesQ.isError,
          items: normChild(industryClassesQ.data?.industry_classes, "industry_group_id", "industry_group"),
        },
      ]}
      levels={levels}
    />
  );
}

export default function HierarchyRoute({ hierarchyId }: { hierarchyId: HierarchyId }) {
  return hierarchyId === "occupation" ? <OccupationRoute /> : <IndustryRoute />;
}
