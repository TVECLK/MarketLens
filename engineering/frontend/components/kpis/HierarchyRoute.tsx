"use client";

import { useState } from "react";
import {
  useMajorGroups, useCreateMajorGroup, useUpdateMajorGroup, useDeleteMajorGroup,
  useSubMajorGroups, useCreateSubMajorGroup, useUpdateSubMajorGroup, useDeleteSubMajorGroup,
  useMinorGroups, useCreateMinorGroup, useUpdateMinorGroup, useDeleteMinorGroup,
  useUnitGroups, useCreateUnitGroup, useUpdateUnitGroup, useDeleteUnitGroup,
  useOccupationGroups, useCreateOccupationGroup, useUpdateOccupationGroup, useDeleteOccupationGroup,

  useIndustrySectors, useCreateIndustrySector, useUpdateIndustrySector, useDeleteIndustrySector,
  useIndustryDivisions, useCreateIndustryDivision, useUpdateIndustryDivision, useDeleteIndustryDivision,
  useIndustryGroups, useCreateIndustryGroup, useUpdateIndustryGroup, useDeleteIndustryGroup,
  useIndustryClasses, useCreateIndustryClass, useUpdateIndustryClass, useDeleteIndustryClass,
  useIndustrySubclasses, useCreateIndustrySubclass, useUpdateIndustrySubclass, useDeleteIndustrySubclass,
} from "@/hooks/use-kpi";
import type { HierarchyId } from "@/lib/kpi-config";
import { normTop, normChild } from "@/lib/kpi-normalizers";
import { clampPage } from "@/lib/pagination";
import type { HierarchyLevelBundle } from "@/types/kpi-admin";
import HierarchyLevelView from "@/components/kpis/HierarchyLevelView";

const PAGE_SIZE = 20;

// All five levels of a hierarchy are loaded together: the active level needs
// the level above for its parent picker, and the delete warning walks every
// level below to count cascaded rows.

function OccupationRoute({ levelIndex }: { levelIndex: number }) {
  const [occGroupPage, setOccGroupPage] = useState(1);

  const majorGroupsQ = useMajorGroups();
  const createMajorGroup = useCreateMajorGroup();
  const updateMajorGroup = useUpdateMajorGroup();
  const deleteMajorGroup = useDeleteMajorGroup();

  const subMajorGroupsQ = useSubMajorGroups();
  const createSubMajorGroup = useCreateSubMajorGroup();
  const updateSubMajorGroup = useUpdateSubMajorGroup();
  const deleteSubMajorGroup = useDeleteSubMajorGroup();

  const minorGroupsQ = useMinorGroups();
  const createMinorGroup = useCreateMinorGroup();
  const updateMinorGroup = useUpdateMinorGroup();
  const deleteMinorGroup = useDeleteMinorGroup();

  const unitGroupsQ = useUnitGroups();
  const createUnitGroup = useCreateUnitGroup();
  const updateUnitGroup = useUpdateUnitGroup();
  const deleteUnitGroup = useDeleteUnitGroup();

  const occupationGroupsQ = useOccupationGroups(PAGE_SIZE, (occGroupPage - 1) * PAGE_SIZE);
  // After a delete empties the last page, step back to the new last page.
  const clampedOccGroupPage = clampPage(occGroupPage, occupationGroupsQ.data?.total, PAGE_SIZE);
  if (clampedOccGroupPage !== occGroupPage) setOccGroupPage(clampedOccGroupPage);
  const createOccupationGroup = useCreateOccupationGroup();
  const updateOccupationGroup = useUpdateOccupationGroup();
  const deleteOccupationGroup = useDeleteOccupationGroup();

  const levels: HierarchyLevelBundle[] = [
    {
      isLoading: majorGroupsQ.isLoading,
      isError: majorGroupsQ.isError,
      items: normTop(majorGroupsQ.data?.major_groups),
      createMut: createMajorGroup,
      updateMut: updateMajorGroup,
      deleteMut: deleteMajorGroup,
      parentField: null,
    },
    {
      isLoading: subMajorGroupsQ.isLoading,
      isError: subMajorGroupsQ.isError,
      items: normChild(subMajorGroupsQ.data?.sub_major_groups, "major_group_id", "major_group"),
      createMut: createSubMajorGroup,
      updateMut: updateSubMajorGroup,
      deleteMut: deleteSubMajorGroup,
      parentField: "major_group_id",
    },
    {
      isLoading: minorGroupsQ.isLoading,
      isError: minorGroupsQ.isError,
      items: normChild(minorGroupsQ.data?.minor_groups, "sub_major_group_id", "sub_major_group"),
      createMut: createMinorGroup,
      updateMut: updateMinorGroup,
      deleteMut: deleteMinorGroup,
      parentField: "sub_major_group_id",
    },
    {
      isLoading: unitGroupsQ.isLoading,
      isError: unitGroupsQ.isError,
      items: normChild(unitGroupsQ.data?.unit_groups, "minor_group_id", "minor_group"),
      createMut: createUnitGroup,
      updateMut: updateUnitGroup,
      deleteMut: deleteUnitGroup,
      parentField: "minor_group_id",
    },
    {
      isLoading: occupationGroupsQ.isLoading,
      isError: occupationGroupsQ.isError,
      items: normChild(occupationGroupsQ.data?.occupation_groups, "unit_group_id", "unit_group"),
      createMut: createOccupationGroup,
      updateMut: updateOccupationGroup,
      deleteMut: deleteOccupationGroup,
      parentField: "unit_group_id",
      total: occupationGroupsQ.data?.total,
      page: occGroupPage,
      setPage: setOccGroupPage,
    },
  ];

  return (
    <HierarchyLevelView key={levelIndex} hierarchyId="occupation" levelIndex={levelIndex} levels={levels} pageSize={PAGE_SIZE} />
  );
}

function IndustryRoute({ levelIndex }: { levelIndex: number }) {
  const [indSubclassPage, setIndSubclassPage] = useState(1);

  const industrySectorsQ = useIndustrySectors();
  const createIndustrySector = useCreateIndustrySector();
  const updateIndustrySector = useUpdateIndustrySector();
  const deleteIndustrySector = useDeleteIndustrySector();

  const industryDivisionsQ = useIndustryDivisions();
  const createIndustryDivision = useCreateIndustryDivision();
  const updateIndustryDivision = useUpdateIndustryDivision();
  const deleteIndustryDivision = useDeleteIndustryDivision();

  const industryGroupsQ = useIndustryGroups();
  const createIndustryGroup = useCreateIndustryGroup();
  const updateIndustryGroup = useUpdateIndustryGroup();
  const deleteIndustryGroup = useDeleteIndustryGroup();

  const industryClassesQ = useIndustryClasses();
  const createIndustryClass = useCreateIndustryClass();
  const updateIndustryClass = useUpdateIndustryClass();
  const deleteIndustryClass = useDeleteIndustryClass();

  const industrySubclassesQ = useIndustrySubclasses(PAGE_SIZE, (indSubclassPage - 1) * PAGE_SIZE);
  // After a delete empties the last page, step back to the new last page.
  const clampedIndSubclassPage = clampPage(indSubclassPage, industrySubclassesQ.data?.total, PAGE_SIZE);
  if (clampedIndSubclassPage !== indSubclassPage) setIndSubclassPage(clampedIndSubclassPage);
  const createIndustrySubclass = useCreateIndustrySubclass();
  const updateIndustrySubclass = useUpdateIndustrySubclass();
  const deleteIndustrySubclass = useDeleteIndustrySubclass();

  const levels: HierarchyLevelBundle[] = [
    {
      isLoading: industrySectorsQ.isLoading,
      isError: industrySectorsQ.isError,
      items: normTop(industrySectorsQ.data?.industry_sectors),
      createMut: createIndustrySector,
      updateMut: updateIndustrySector,
      deleteMut: deleteIndustrySector,
      parentField: null,
    },
    {
      isLoading: industryDivisionsQ.isLoading,
      isError: industryDivisionsQ.isError,
      items: normChild(industryDivisionsQ.data?.industry_divisions, "industry_sector_id", "industry_sector"),
      createMut: createIndustryDivision,
      updateMut: updateIndustryDivision,
      deleteMut: deleteIndustryDivision,
      parentField: "industry_sector_id",
    },
    {
      isLoading: industryGroupsQ.isLoading,
      isError: industryGroupsQ.isError,
      items: normChild(industryGroupsQ.data?.industry_groups, "industry_division_id", "industry_division"),
      createMut: createIndustryGroup,
      updateMut: updateIndustryGroup,
      deleteMut: deleteIndustryGroup,
      parentField: "industry_division_id",
    },
    {
      isLoading: industryClassesQ.isLoading,
      isError: industryClassesQ.isError,
      items: normChild(industryClassesQ.data?.industry_classes, "industry_group_id", "industry_group"),
      createMut: createIndustryClass,
      updateMut: updateIndustryClass,
      deleteMut: deleteIndustryClass,
      parentField: "industry_group_id",
    },
    {
      isLoading: industrySubclassesQ.isLoading,
      isError: industrySubclassesQ.isError,
      items: normChild(industrySubclassesQ.data?.industry_subclasses, "industry_class_id", "industry_class"),
      createMut: createIndustrySubclass,
      updateMut: updateIndustrySubclass,
      deleteMut: deleteIndustrySubclass,
      parentField: "industry_class_id",
      total: industrySubclassesQ.data?.total,
      page: indSubclassPage,
      setPage: setIndSubclassPage,
    },
  ];

  return (
    <HierarchyLevelView key={levelIndex} hierarchyId="industry" levelIndex={levelIndex} levels={levels} pageSize={PAGE_SIZE} />
  );
}

export default function HierarchyRoute({
  hierarchyId,
  levelIndex,
}: {
  hierarchyId: HierarchyId;
  levelIndex: number;
}) {
  return hierarchyId === "occupation" ? (
    <OccupationRoute levelIndex={levelIndex} />
  ) : (
    <IndustryRoute levelIndex={levelIndex} />
  );
}
