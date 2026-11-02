"use client";

import type { ComponentType } from "react";
import {
  useEmploymentSectors, useCreateEmploymentSector, useUpdateEmploymentSector, useDeleteEmploymentSector,
  useExperiences, useCreateExperience, useUpdateExperience, useDeleteExperience,
  useEducationLevels, useCreateEducationLevel, useUpdateEducationLevel, useDeleteEducationLevel,
  useFormalities, useCreateFormality, useUpdateFormality, useDeleteFormality,
  useGenders, useCreateGender, useUpdateGender, useDeleteGender,
  useVocationalEducations, useCreateVocationalEducation, useUpdateVocationalEducation, useDeleteVocationalEducation,
} from "@/hooks/use-kpi";
import type { KPICategoryKey } from "@/lib/kpi-config";
import type { FlatBundle, FlatItem } from "@/types/kpi-admin";
import FlatCategoryView from "@/components/kpis/FlatCategoryView";

interface ListQueryLike<TData> {
  data: TData | undefined;
  isLoading: boolean;
  isError: boolean;
}

// Builds one view component per category
function flatView<TData>(
  category: KPICategoryKey,
  labelField: string,
  useList: () => ListQueryLike<TData>,
  toItems: (data: TData | undefined) => FlatItem[],
  useCreate: () => FlatBundle["createMut"],
  useUpdate: () => FlatBundle["updateMut"],
  useDelete: () => FlatBundle["deleteMut"],
): ComponentType {
  function View() {
    const listQ = useList();
    const createMut = useCreate();
    const updateMut = useUpdate();
    const deleteMut = useDelete();
    const bundle: FlatBundle = {
      isLoading: listQ.isLoading,
      isError: listQ.isError,
      items: toItems(listQ.data),
      createMut,
      updateMut,
      deleteMut,
      labelField,
    };
    return <FlatCategoryView category={category} bundle={bundle} />;
  }
  View.displayName = `FlatView(${category})`;
  return View;
}

const FLAT_VIEWS: Record<KPICategoryKey, ComponentType> = {
  employmentSector: flatView(
    "employmentSector", "sector", useEmploymentSectors,
    (d) => (d?.employment_sectors ?? []).map((e) => ({ id: e.id, label: e.sector })),
    useCreateEmploymentSector, useUpdateEmploymentSector, useDeleteEmploymentSector,
  ),
  experience: flatView(
    "experience", "name", useExperiences,
    (d) => (d?.experiences ?? []).map((e) => ({ id: e.id, label: e.name })),
    useCreateExperience, useUpdateExperience, useDeleteExperience,
  ),
  educationLevel: flatView(
    "educationLevel", "level", useEducationLevels,
    (d) => (d?.education_levels ?? []).map((e) => ({ id: e.id, label: e.level })),
    useCreateEducationLevel, useUpdateEducationLevel, useDeleteEducationLevel,
  ),
  formalInformal: flatView(
    "formalInformal", "formality_type", useFormalities,
    (d) => (d?.formalities ?? []).map((f) => ({ id: f.id, label: f.formality_type })),
    useCreateFormality, useUpdateFormality, useDeleteFormality,
  ),
  gender: flatView(
    "gender", "gender_type", useGenders,
    (d) => (d?.genders ?? []).map((g) => ({ id: g.id, label: g.gender_type })),
    useCreateGender, useUpdateGender, useDeleteGender,
  ),
  vocationalEducation: flatView(
    "vocationalEducation", "level", useVocationalEducations,
    (d) => (d?.vocational_educations ?? []).map((v) => ({ id: v.id, label: v.level })),
    useCreateVocationalEducation, useUpdateVocationalEducation, useDeleteVocationalEducation,
  ),
};

export default function FlatCategoryRoute({ category }: { category: KPICategoryKey }) {
  const View = FLAT_VIEWS[category];
  return <View />;
}
