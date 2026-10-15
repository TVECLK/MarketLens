"use client";

import { useState, type FormEvent } from "react";
import Header from "@/components/layout/Header";
import type { HierarchyRow } from "@/types/kpi";
import {
  useEmploymentSectors, useCreateEmploymentSector, useUpdateEmploymentSector, useDeleteEmploymentSector,
  useExperiences, useCreateExperience, useUpdateExperience, useDeleteExperience,
  useEducationLevels, useCreateEducationLevel, useUpdateEducationLevel, useDeleteEducationLevel,
  useFormalities, useCreateFormality, useUpdateFormality, useDeleteFormality,
  useGenders, useCreateGender, useUpdateGender, useDeleteGender,
  useVocationalEducations, useCreateVocationalEducation, useUpdateVocationalEducation, useDeleteVocationalEducation,

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
import {
  KPI_CATEGORIES,
  HIERARCHIES,
  ACCENT_CLASSES,
  type KPICategoryKey,
  type HierarchyId,
} from "@/lib/kpi-config";
import { normTop, normChild } from "@/lib/kpi-normalizers";
import type { FlatBundle, HierarchyLevelBundle, SectionKey } from "@/types/kpi-admin";
import CategoryNav from "@/components/kpis/CategoryNav";
import FlatCategoryPanel from "@/components/kpis/FlatCategoryPanel";
import HierarchyLevelPanel from "@/components/kpis/HierarchyLevelPanel";
import ConfirmDeleteModal from "@/components/kpis/ConfirmDeleteModal";

function errMsg(err: unknown, fallback: string): string {
  return err instanceof Error ? err.message : fallback;
}

export default function ManageKpisPage() {
  const [activeSection, setActiveSection] = useState<SectionKey>("flat");

  // FLAT KPI HOOKS — called unconditionally, all 6, every render (Rules of
  // Hooks). Which one is "active" is decided afterwards via flatBundles.
  const employmentSectorsQ = useEmploymentSectors();
  const createEmploymentSector = useCreateEmploymentSector();
  const updateEmploymentSector = useUpdateEmploymentSector();
  const deleteEmploymentSector = useDeleteEmploymentSector();

  const experiencesQ = useExperiences();
  const createExperience = useCreateExperience();
  const updateExperience = useUpdateExperience();
  const deleteExperience = useDeleteExperience();

  const educationLevelsQ = useEducationLevels();
  const createEducationLevel = useCreateEducationLevel();
  const updateEducationLevel = useUpdateEducationLevel();
  const deleteEducationLevel = useDeleteEducationLevel();

  const formalitiesQ = useFormalities();
  const createFormality = useCreateFormality();
  const updateFormality = useUpdateFormality();
  const deleteFormality = useDeleteFormality();

  const gendersQ = useGenders();
  const createGender = useCreateGender();
  const updateGender = useUpdateGender();
  const deleteGender = useDeleteGender();

  const vocationalEducationsQ = useVocationalEducations();
  const createVocationalEducation = useCreateVocationalEducation();
  const updateVocationalEducation = useUpdateVocationalEducation();
  const deleteVocationalEducation = useDeleteVocationalEducation();

  const flatBundles: Record<KPICategoryKey, FlatBundle> = {
    employmentSector: {
      isLoading: employmentSectorsQ.isLoading,
      isError: employmentSectorsQ.isError,
      items: (employmentSectorsQ.data?.employment_sectors ?? []).map((e) => ({ id: e.id, label: e.sector })),
      createMut: createEmploymentSector,
      updateMut: updateEmploymentSector,
      deleteMut: deleteEmploymentSector,
      labelField: "sector",
    },
    experience: {
      isLoading: experiencesQ.isLoading,
      isError: experiencesQ.isError,
      items: (experiencesQ.data?.experiences ?? []).map((e) => ({ id: e.id, label: e.name })),
      createMut: createExperience,
      updateMut: updateExperience,
      deleteMut: deleteExperience,
      labelField: "name",
    },
    educationLevel: {
      isLoading: educationLevelsQ.isLoading,
      isError: educationLevelsQ.isError,
      items: (educationLevelsQ.data?.education_levels ?? []).map((e) => ({ id: e.id, label: e.level })),
      createMut: createEducationLevel,
      updateMut: updateEducationLevel,
      deleteMut: deleteEducationLevel,
      labelField: "level",
    },
    formalInformal: {
      isLoading: formalitiesQ.isLoading,
      isError: formalitiesQ.isError,
      items: (formalitiesQ.data?.formalities ?? []).map((f) => ({ id: f.id, label: f.formality_type })),
      createMut: createFormality,
      updateMut: updateFormality,
      deleteMut: deleteFormality,
      labelField: "formality_type",
    },
    gender: {
      isLoading: gendersQ.isLoading,
      isError: gendersQ.isError,
      items: (gendersQ.data?.genders ?? []).map((g) => ({ id: g.id, label: g.gender_type })),
      createMut: createGender,
      updateMut: updateGender,
      deleteMut: deleteGender,
      labelField: "gender_type",
    },
    vocationalEducation: {
      isLoading: vocationalEducationsQ.isLoading,
      isError: vocationalEducationsQ.isError,
      items: (vocationalEducationsQ.data?.vocational_educations ?? []).map((v) => ({ id: v.id, label: v.level })),
      createMut: createVocationalEducation,
      updateMut: updateVocationalEducation,
      deleteMut: deleteVocationalEducation,
      labelField: "level",
    },
  };

  const [activeCategory, setActiveCategory] = useState<KPICategoryKey>("employmentSector");
  const config = KPI_CATEGORIES[activeCategory];
  const accent = ACCENT_CLASSES[config.accent];
  const bundle = flatBundles[activeCategory];

  const [newName, setNewName] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editError, setEditError] = useState<string | null>(null);

  const [deleteTarget, setDeleteTarget] = useState<{ id: number; label: string } | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  function flashToast(msg: string) {
    setToast(msg);
    setTimeout(() => setToast(null), 2200);
  }

  function switchCategory(cat: KPICategoryKey) {
    setActiveSection("flat");
    setActiveCategory(cat);
    setEditingId(null);
    setEditError(null);
    setFormError(null);
    setNewName("");
  }

  // CREATE
  function handleAdd(e: FormEvent) {
    e.preventDefault();
    const trimmed = newName.trim();
    if (!trimmed) {
      setFormError(`${config.itemLabel} name is required.`);
      return;
    }
    setFormError(null);
    bundle.createMut.mutate(
      { [bundle.labelField]: trimmed },
      {
        onSuccess: () => {
          setNewName("");
          flashToast(`Added "${trimmed}"`);
        },
        onError: (err) => setFormError(errMsg(err, "Failed to create item.")),
      }
    );
  }

  // UPDATE
  function startEdit(item: { id: number; label: string }) {
    setEditingId(item.id);
    setEditName(item.label);
    setEditError(null);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditError(null);
  }

  function saveEdit(id: number) {
    const trimmed = editName.trim();
    if (!trimmed) {
      setEditError(`${config.itemLabel} name is required.`);
      return;
    }
    bundle.updateMut.mutate(
      { id, payload: { [bundle.labelField]: trimmed } },
      {
        onSuccess: () => {
          setEditingId(null);
          setEditError(null);
          flashToast(`Updated "${trimmed}"`);
        },
        onError: (err) => setEditError(errMsg(err, "Failed to update item.")),
      }
    );
  }

  // DELETE
  function handleDelete() {
    if (!deleteTarget) return;
    const target = deleteTarget;
    bundle.deleteMut.mutate(target.id, {
      onSuccess: () => {
        flashToast(`Deleted "${target.label}"`);
        setDeleteTarget(null);
      },
      onError: (err) => {
        flashToast(errMsg(err, "Failed to delete item."));
        setDeleteTarget(null);
      },
    });
  }


  // HIERARCHICAL KPI HOOKS — all 10 levels called unconditionally, every
  // render. Occupation branch:
  const PAGE_SIZE = 20;
  const [occGroupPage, setOccGroupPage] = useState(1);
  const [indSubclassPage, setIndSubclassPage] = useState(1);

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
  const createOccupationGroup = useCreateOccupationGroup();
  const updateOccupationGroup = useUpdateOccupationGroup();
  const deleteOccupationGroup = useDeleteOccupationGroup();

  // Industry branch:
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
  const createIndustrySubclass = useCreateIndustrySubclass();
  const updateIndustrySubclass = useUpdateIndustrySubclass();
  const deleteIndustrySubclass = useDeleteIndustrySubclass();

  const hierarchyBundles: Record<HierarchyId, HierarchyLevelBundle[]> = {
    occupation: [
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
    ],
    industry: [
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
    ],
  };

  const [hActiveLevelIndex, setHActiveLevelIndex] = useState(0);

  const [hNewName, setHNewName] = useState("");
  const [hNewCode, setHNewCode] = useState("");
  const [hNewParentId, setHNewParentId] = useState("");
  const [hFormError, setHFormError] = useState<string | null>(null);

  const [hEditingId, setHEditingId] = useState<number | null>(null);
  const [hEditName, setHEditName] = useState("");
  const [hEditCode, setHEditCode] = useState("");
  const [hEditParentId, setHEditParentId] = useState("");
  const [hEditError, setHEditError] = useState<string | null>(null);

  const [hDeleteTarget, setHDeleteTarget] = useState<HierarchyRow | null>(null);

  const isHierarchySection = activeSection === "occupation" || activeSection === "industry";
  const hierarchyId: HierarchyId | null = isHierarchySection ? (activeSection as HierarchyId) : null;
  const hierarchyDef = hierarchyId ? HIERARCHIES[hierarchyId] : null;
  const hAccent = hierarchyDef ? ACCENT_CLASSES[hierarchyDef.accent] : ACCENT_CLASSES.rose;
  const hLevelDef = hierarchyDef ? hierarchyDef.levels[hActiveLevelIndex] : null;
  const hParentLevelDef = hierarchyDef && hActiveLevelIndex > 0 ? hierarchyDef.levels[hActiveLevelIndex - 1] : null;

  const hLevelBundle = hierarchyId ? hierarchyBundles[hierarchyId][hActiveLevelIndex] : null;
  const hParentBundle = hierarchyId && hActiveLevelIndex > 0 ? hierarchyBundles[hierarchyId][hActiveLevelIndex - 1] : null;
  const hParentItems = hParentBundle?.items ?? [];

  function resetHForms() {
    setHEditingId(null);
    setHEditError(null);
    setHFormError(null);
    setHNewName("");
    setHNewCode("");
    setHNewParentId("");
  }

  function switchHierarchy(id: HierarchyId) {
    setActiveSection(id);
    setHActiveLevelIndex(0);
    resetHForms();
  }

  function switchHLevel(index: number) {
    setHActiveLevelIndex(index);
    resetHForms();
  }

  // CREATE
  function handleHAdd(e: FormEvent) {
    e.preventDefault();
    if (!hierarchyId || !hLevelDef || !hLevelBundle) return;

    const trimmedName = hNewName.trim();
    const trimmedCode = hNewCode.trim();

    if (!trimmedName) {
      setHFormError(`${hLevelDef.itemLabel} name is required.`);
      return;
    }
    if (!trimmedCode) {
      setHFormError(`${hLevelDef.itemLabel} code is required.`);
      return;
    }
    if (hParentLevelDef && !hNewParentId) {
      setHFormError(`Select a parent ${hParentLevelDef.itemLabel}.`);
      return;
    }
    setHFormError(null);

    const payload: Record<string, unknown> = { name: trimmedName, code: trimmedCode };
    if (hLevelBundle.parentField) {
      payload[hLevelBundle.parentField] = Number(hNewParentId);
    }

    hLevelBundle.createMut.mutate(payload, {
      onSuccess: () => {
        setHNewName("");
        setHNewCode("");
        setHNewParentId("");
        flashToast(`Added "${trimmedName}"`);
      },
      onError: (err) => setHFormError(errMsg(err, "Failed to create item.")),
    });
  }

  // UPDATE
  function startHEdit(item: HierarchyRow) {
    setHEditingId(item.id);
    setHEditName(item.name);
    setHEditCode(item.code);
    setHEditParentId(item.parentId != null ? String(item.parentId) : "");
    setHEditError(null);
  }

  function cancelHEdit() {
    setHEditingId(null);
    setHEditError(null);
  }

  function saveHEdit(id: number) {
    if (!hierarchyId || !hLevelDef || !hLevelBundle) return;

    const trimmedName = hEditName.trim();
    const trimmedCode = hEditCode.trim();

    if (!trimmedName) {
      setHEditError(`${hLevelDef.itemLabel} name is required.`);
      return;
    }
    if (!trimmedCode) {
      setHEditError(`${hLevelDef.itemLabel} code is required.`);
      return;
    }
    if (hParentLevelDef && !hEditParentId) {
      setHEditError(`Select a parent ${hParentLevelDef.itemLabel}.`);
      return;
    }

    const payload: Record<string, unknown> = { name: trimmedName, code: trimmedCode };
    if (hLevelBundle.parentField) {
      payload[hLevelBundle.parentField] = Number(hEditParentId);
    }

    hLevelBundle.updateMut.mutate(
      { id, payload },
      {
        onSuccess: () => {
          setHEditingId(null);
          setHEditError(null);
          flashToast(`Updated "${trimmedName}"`);
        },
        onError: (err) => setHEditError(errMsg(err, "Failed to update item.")),
      }
    );
  }

  // DELETE
  // Note: the DB has ON DELETE CASCADE on every level's parent_id foreign
  // key, so deleting a row here removes all its descendants server-side
  // automatically — the delete hooks (see hooks/use-kpi.ts) already
  // invalidate every level's cache for this hierarchy as a result.
  function handleHDelete() {
    if (!hLevelBundle || !hDeleteTarget) return;
    const target = hDeleteTarget;
    hLevelBundle.deleteMut.mutate(target.id, {
      onSuccess: () => {
        flashToast(`Deleted "${target.name}"`);
        setHDeleteTarget(null);
      },
      onError: (err) => {
        flashToast(errMsg(err, "Failed to delete item."));
        setHDeleteTarget(null);
      },
    });
  }

  function countHierarchyDescendants(id: HierarchyId, levelIndex: number, targetId: number): number {
    const levels = hierarchyBundles[id];
    if (levelIndex === levels.length - 1) return 0;
    const childItems = levels[levelIndex + 1].items.filter((c) => c.parentId === targetId);
    return childItems.reduce(
      (sum, child) => sum + 1 + countHierarchyDescendants(id, levelIndex + 1, child.id),
      0
    );
  }

  const hDeleteDescendantCount =
    hierarchyId && hDeleteTarget ? countHierarchyDescendants(hierarchyId, hActiveLevelIndex, hDeleteTarget.id) : 0;

  const flatCounts: Record<KPICategoryKey, number> = {
    employmentSector: flatBundles.employmentSector.items.length,
    experience: flatBundles.experience.items.length,
    educationLevel: flatBundles.educationLevel.items.length,
    formalInformal: flatBundles.formalInformal.items.length,
    gender: flatBundles.gender.items.length,
    vocationalEducation: flatBundles.vocationalEducation.items.length,
  };

  const levelCounts = hierarchyId ? hierarchyBundles[hierarchyId].map((lvl) => lvl.items.length) : [];

  return (
    <div className="min-h-screen bg-gray-50 text-gray-800 font-sans">
      <Header
        title="KPI Reference Data Manager"
        subtitle="Create, edit, and remove the reference values that power the labour market dashboard charts"
      />

      <div className="flex">
        <CategoryNav
          activeSection={activeSection}
          activeCategory={activeCategory}
          flatCounts={flatCounts}
          onSwitchCategory={switchCategory}
          onSwitchHierarchy={switchHierarchy}
        />

        {activeSection === "flat" ? (
          <FlatCategoryPanel
            config={config}
            accent={accent}
            bundle={bundle}
            newName={newName}
            onNewNameChange={setNewName}
            formError={formError}
            onSubmit={handleAdd}
            editingId={editingId}
            editName={editName}
            onEditNameChange={setEditName}
            editError={editError}
            onStartEdit={startEdit}
            onCancelEdit={cancelEdit}
            onSaveEdit={saveEdit}
            onRequestDelete={setDeleteTarget}
          />
        ) : hierarchyDef && hLevelBundle ? (
          <HierarchyLevelPanel
            hierarchyDef={hierarchyDef}
            hAccent={hAccent}
            activeLevelIndex={hActiveLevelIndex}
            levelCounts={levelCounts}
            onSwitchLevel={switchHLevel}
            pageSize={PAGE_SIZE}
            levelBundle={hLevelBundle}
            parentItems={hParentItems}
            newName={hNewName}
            onNewNameChange={setHNewName}
            newCode={hNewCode}
            onNewCodeChange={setHNewCode}
            newParentId={hNewParentId}
            onNewParentIdChange={setHNewParentId}
            formError={hFormError}
            onSubmit={handleHAdd}
            editingId={hEditingId}
            editName={hEditName}
            onEditNameChange={setHEditName}
            editCode={hEditCode}
            onEditCodeChange={setHEditCode}
            editParentId={hEditParentId}
            onEditParentIdChange={setHEditParentId}
            editError={hEditError}
            onStartEdit={startHEdit}
            onCancelEdit={cancelHEdit}
            onSaveEdit={saveHEdit}
            onRequestDelete={setHDeleteTarget}
          />
        ) : null}
      </div>

      {deleteTarget && (
        <ConfirmDeleteModal
          title={`Delete ${config.itemLabel}?`}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={handleDelete}
          isPending={bundle.deleteMut.isPending}
        >
          This will permanently remove{" "}
          <span className="font-bold text-gray-800">&quot;{deleteTarget.label}&quot;</span> from{" "}
          {config.title}. This action can&apos;t be undone.
        </ConfirmDeleteModal>
      )}

      {hDeleteTarget && hLevelBundle && hLevelDef && (
        <ConfirmDeleteModal
          title={`Delete ${hLevelDef.itemLabel}?`}
          onCancel={() => setHDeleteTarget(null)}
          onConfirm={handleHDelete}
          isPending={hLevelBundle.deleteMut.isPending}
        >
          <p>
            This will permanently remove{" "}
            <span className="font-bold text-gray-800">
              &quot;{hDeleteTarget.code} — {hDeleteTarget.name}&quot;
            </span>
            . This action can&apos;t be undone.
          </p>
          {hDeleteDescendantCount > 0 && (
            <p className="font-bold text-red-500 mt-2">
              It also has {hDeleteDescendantCount} item{hDeleteDescendantCount === 1 ? "" : "s"} nested
              underneath it across the lower levels — those will be deleted too.
            </p>
          )}
        </ConfirmDeleteModal>
      )}

      {toast && (
        <div className="fixed bottom-6 right-6 bg-gray-900 text-white text-xs font-bold px-4 py-3 rounded-xl shadow-lg z-50">
          {toast}
        </div>
      )}
    </div>
  );
}
