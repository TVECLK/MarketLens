"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type {
  EmploymentSector, EmploymentSectorListResponse, EmploymentSectorPayload,
  Experience, ExperienceListResponse, ExperiencePayload,
  EducationLevel, EducationLevelListResponse, EducationLevelPayload,
  Formality, FormalityListResponse, FormalityPayload,
  Gender, GenderListResponse, GenderPayload,
  VocationalEducation, VocationalEducationListResponse, VocationalEducationPayload,
  MajorGroup, MajorGroupListResponse, MajorGroupPayload,
  SubMajorGroup, SubMajorGroupListResponse, SubMajorGroupPayload,
  MinorGroup, MinorGroupListResponse, MinorGroupPayload,
  UnitGroup, UnitGroupListResponse, UnitGroupPayload,
  OccupationGroup, OccupationGroupListResponse, OccupationGroupPayload,
  IndustrySector, IndustrySectorListResponse, IndustrySectorPayload,
  IndustryDivision, IndustryDivisionListResponse, IndustryDivisionPayload,
  IndustryGroup, IndustryGroupListResponse, IndustryGroupPayload,
  IndustryClass, IndustryClassListResponse, IndustryClassPayload,
  IndustrySubclass, IndustrySubclassListResponse, IndustrySubclassPayload,
} from "@/types/kpi";
import { ENV } from "@/config/env";
import { getApiErrorMessage } from "@/lib/errors";
import { OCCUPATION_QUERY_KEYS, INDUSTRY_QUERY_KEYS } from "@/lib/kpi-config";
import { invalidateAfterKpiChange, relatedListKeys } from "@/lib/query-invalidation";

const BASE = ENV.NEXT_PUBLIC_API_BASE_URL;


function useListQuery<TResponse>(
   key: string,
   endpoint: string,
   params?: Record<string, number | undefined>
 ) {
   const query = new URLSearchParams();
   if (params) {
     Object.entries(params).forEach(([k, v]) => {
       if (v !== undefined) query.set(k, String(v));
     });
   }
   const qs = query.toString();
   const url = qs ? `${endpoint}?${qs}` : endpoint;
  return useQuery<TResponse>({
    queryKey: [key, "list", params ?? {}],
    queryFn: async ({ signal }) => {
      const res = await fetch(url, { signal });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throw new Error(getApiErrorMessage(data, `Failed to fetch ${key}`));
      }
      return data as TResponse;
    },
  });
}

function useCreateMutation<TPayload, TResult>(key: string, endpoint: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: TPayload) => {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throw new Error(getApiErrorMessage(data, `Failed to create ${key}`));
      }
      return data as TResult;
    },
    onSuccess: () => invalidateAfterKpiChange(queryClient, relatedListKeys(key)),
  });
}

function useUpdateMutation<TPayload, TResult>(key: string, endpoint: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, payload }: { id: number; payload: Partial<TPayload> }) => {
      const res = await fetch(`${endpoint}/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throw new Error(getApiErrorMessage(data, `Failed to update ${key}`));
      }
      return data as TResult;
    },
    onSuccess: () => invalidateAfterKpiChange(queryClient, relatedListKeys(key)),
  });
}

/** Plain delete: invalidates only this entity's own admin list (plus the
 * dashboard queries). Used for flat KPIs (no children, so no cascade concern). */
function useDeleteMutation(key: string, endpoint: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      const res = await fetch(`${endpoint}/${id}`, { method: "DELETE" });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throw new Error(getApiErrorMessage(data, `Failed to delete ${key}`));
      }
      return data;
    },
    onSuccess: () => invalidateAfterKpiChange(queryClient, [key]),
  });
}

/** Cascade-aware delete for hierarchy levels: the backend soft-deletes the
 * row (sets deleted_at) and, in the same transaction, soft-deletes all its
 * descendants itself, since soft delete doesn't trigger any DB-level cascade.
 * The client just needs to invalidate every level's cache in that hierarchy
 * (not only the level that was deleted), since we don't track which
 * lower-level rows were hidden as a result. */
function useHierarchyDeleteMutation(key: string, endpoint: string, relatedKeys: string[]) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      const res = await fetch(`${endpoint}/${id}`, { method: "DELETE" });
      const data = await res.json().catch(() => null);
      if (!res.ok) {
        throw new Error(getApiErrorMessage(data, `Failed to delete ${key}`));
      }
      return data;
    },
    onSuccess: () => invalidateAfterKpiChange(queryClient, relatedKeys),
  });
}

// EMPLOYMENTSECTOR ("employment-sectors")

const EMPLOYMENTSECTORS_URL = `${BASE}/employment-sectors`;

export function useEmploymentSectors() {
  return useListQuery<EmploymentSectorListResponse>("employment-sectors", EMPLOYMENTSECTORS_URL);
}

export function useCreateEmploymentSector() {
  return useCreateMutation<EmploymentSectorPayload, EmploymentSector>("employment-sectors", EMPLOYMENTSECTORS_URL);
}

export function useUpdateEmploymentSector() {
  return useUpdateMutation<EmploymentSectorPayload, EmploymentSector>("employment-sectors", EMPLOYMENTSECTORS_URL);
}

export function useDeleteEmploymentSector() {
  return useDeleteMutation("employment-sectors", EMPLOYMENTSECTORS_URL);
}


// EXPERIENCE ("experiences")

const EXPERIENCES_URL = `${BASE}/experiences`;

export function useExperiences() {
  return useListQuery<ExperienceListResponse>("experiences", EXPERIENCES_URL);
}

export function useCreateExperience() {
  return useCreateMutation<ExperiencePayload, Experience>("experiences", EXPERIENCES_URL);
}

export function useUpdateExperience() {
  return useUpdateMutation<ExperiencePayload, Experience>("experiences", EXPERIENCES_URL);
}

export function useDeleteExperience() {
  return useDeleteMutation("experiences", EXPERIENCES_URL);
}


// EDUCATIONLEVEL ("education-levels")

const EDUCATIONLEVELS_URL = `${BASE}/education-levels`;

export function useEducationLevels() {
  return useListQuery<EducationLevelListResponse>("education-levels", EDUCATIONLEVELS_URL);
}

export function useCreateEducationLevel() {
  return useCreateMutation<EducationLevelPayload, EducationLevel>("education-levels", EDUCATIONLEVELS_URL);
}

export function useUpdateEducationLevel() {
  return useUpdateMutation<EducationLevelPayload, EducationLevel>("education-levels", EDUCATIONLEVELS_URL);
}

export function useDeleteEducationLevel() {
  return useDeleteMutation("education-levels", EDUCATIONLEVELS_URL);
}


// FORMALITY ("formalities")
const FORMALITIES_URL = `${BASE}/formalities`;

export function useFormalities() {
  return useListQuery<FormalityListResponse>("formalities", FORMALITIES_URL);
}

export function useCreateFormality() {
  return useCreateMutation<FormalityPayload, Formality>("formalities", FORMALITIES_URL);
}

export function useUpdateFormality() {
  return useUpdateMutation<FormalityPayload, Formality>("formalities", FORMALITIES_URL);
}

export function useDeleteFormality() {
  return useDeleteMutation("formalities", FORMALITIES_URL);
}


// GENDER ("genders")

const GENDERS_URL = `${BASE}/genders`;

export function useGenders() {
  return useListQuery<GenderListResponse>("genders", GENDERS_URL);
}

export function useCreateGender() {
  return useCreateMutation<GenderPayload, Gender>("genders", GENDERS_URL);
}

export function useUpdateGender() {
  return useUpdateMutation<GenderPayload, Gender>("genders", GENDERS_URL);
}

export function useDeleteGender() {
  return useDeleteMutation("genders", GENDERS_URL);
}


// VOCATIONALEDUCATION ("vocational-educations")

const VOCATIONALEDUCATIONS_URL = `${BASE}/vocational-educations`;

export function useVocationalEducations() {
  return useListQuery<VocationalEducationListResponse>("vocational-educations", VOCATIONALEDUCATIONS_URL);
}

export function useCreateVocationalEducation() {
  return useCreateMutation<VocationalEducationPayload, VocationalEducation>("vocational-educations", VOCATIONALEDUCATIONS_URL);
}

export function useUpdateVocationalEducation() {
  return useUpdateMutation<VocationalEducationPayload, VocationalEducation>("vocational-educations", VOCATIONALEDUCATIONS_URL);
}

export function useDeleteVocationalEducation() {
  return useDeleteMutation("vocational-educations", VOCATIONALEDUCATIONS_URL);
}


// MAJORGROUP ("major-groups") - occupation hierarchy level

const MAJORGROUPS_URL = `${BASE}/occupation/major-groups`;

export function useMajorGroups() {
  return useListQuery<MajorGroupListResponse>("major-groups", MAJORGROUPS_URL);
}

export function useCreateMajorGroup() {
  return useCreateMutation<MajorGroupPayload, MajorGroup>("major-groups", MAJORGROUPS_URL);
}

export function useUpdateMajorGroup() {
  return useUpdateMutation<MajorGroupPayload, MajorGroup>("major-groups", MAJORGROUPS_URL);
}

export function useDeleteMajorGroup() {
  return useHierarchyDeleteMutation("major-groups", MAJORGROUPS_URL, OCCUPATION_QUERY_KEYS);
}


// SUBMAJORGROUP ("sub-major-groups") - occupation hierarchy level

const SUBMAJORGROUPS_URL = `${BASE}/occupation/sub-major-groups`;

export function useSubMajorGroups() {
  return useListQuery<SubMajorGroupListResponse>("sub-major-groups", SUBMAJORGROUPS_URL);
}

export function useCreateSubMajorGroup() {
  return useCreateMutation<SubMajorGroupPayload, SubMajorGroup>("sub-major-groups", SUBMAJORGROUPS_URL);
}

export function useUpdateSubMajorGroup() {
  return useUpdateMutation<SubMajorGroupPayload, SubMajorGroup>("sub-major-groups", SUBMAJORGROUPS_URL);
}

export function useDeleteSubMajorGroup() {
  return useHierarchyDeleteMutation("sub-major-groups", SUBMAJORGROUPS_URL, OCCUPATION_QUERY_KEYS);
}


// MINORGROUP ("minor-groups") - occupation hierarchy level

const MINORGROUPS_URL = `${BASE}/occupation/minor-groups`;

export function useMinorGroups() {
  return useListQuery<MinorGroupListResponse>("minor-groups", MINORGROUPS_URL);
}

export function useCreateMinorGroup() {
  return useCreateMutation<MinorGroupPayload, MinorGroup>("minor-groups", MINORGROUPS_URL);
}

export function useUpdateMinorGroup() {
  return useUpdateMutation<MinorGroupPayload, MinorGroup>("minor-groups", MINORGROUPS_URL);
}

export function useDeleteMinorGroup() {
  return useHierarchyDeleteMutation("minor-groups", MINORGROUPS_URL, OCCUPATION_QUERY_KEYS);
}


// UNITGROUP ("unit-groups") - occupation hierarchy level

const UNITGROUPS_URL = `${BASE}/occupation/unit-groups`;

export function useUnitGroups() {
  return useListQuery<UnitGroupListResponse>("unit-groups", UNITGROUPS_URL);
}

export function useCreateUnitGroup() {
  return useCreateMutation<UnitGroupPayload, UnitGroup>("unit-groups", UNITGROUPS_URL);
}

export function useUpdateUnitGroup() {
  return useUpdateMutation<UnitGroupPayload, UnitGroup>("unit-groups", UNITGROUPS_URL);
}

export function useDeleteUnitGroup() {
  return useHierarchyDeleteMutation("unit-groups", UNITGROUPS_URL, OCCUPATION_QUERY_KEYS);
}


// OCCUPATIONGROUP ("occupation-groups") - occupation hierarchy level

const OCCUPATIONGROUPS_URL = `${BASE}/occupation/occupation-groups`;

export function useOccupationGroups(limit?: number, offset?: number) {
   return useListQuery<OccupationGroupListResponse>("occupation-groups", OCCUPATIONGROUPS_URL, { limit, offset });
}

export function useCreateOccupationGroup() {
  return useCreateMutation<OccupationGroupPayload, OccupationGroup>("occupation-groups", OCCUPATIONGROUPS_URL);
}

export function useUpdateOccupationGroup() {
  return useUpdateMutation<OccupationGroupPayload, OccupationGroup>("occupation-groups", OCCUPATIONGROUPS_URL);
}

export function useDeleteOccupationGroup() {
  return useHierarchyDeleteMutation("occupation-groups", OCCUPATIONGROUPS_URL, OCCUPATION_QUERY_KEYS);
}


// INDUSTRYSECTOR ("industry-sectors") - industry hierarchy level

const INDUSTRYSECTORS_URL = `${BASE}/industry/industry-sectors`;

export function useIndustrySectors() {
  return useListQuery<IndustrySectorListResponse>("industry-sectors", INDUSTRYSECTORS_URL);
}

export function useCreateIndustrySector() {
  return useCreateMutation<IndustrySectorPayload, IndustrySector>("industry-sectors", INDUSTRYSECTORS_URL);
}

export function useUpdateIndustrySector() {
  return useUpdateMutation<IndustrySectorPayload, IndustrySector>("industry-sectors", INDUSTRYSECTORS_URL);
}

export function useDeleteIndustrySector() {
  return useHierarchyDeleteMutation("industry-sectors", INDUSTRYSECTORS_URL, INDUSTRY_QUERY_KEYS);
}


// INDUSTRYDIVISION ("industry-divisions") - industry hierarchy level

const INDUSTRYDIVISIONS_URL = `${BASE}/industry/industry-divisions`;

export function useIndustryDivisions() {
  return useListQuery<IndustryDivisionListResponse>("industry-divisions", INDUSTRYDIVISIONS_URL);
}

export function useCreateIndustryDivision() {
  return useCreateMutation<IndustryDivisionPayload, IndustryDivision>("industry-divisions", INDUSTRYDIVISIONS_URL);
}

export function useUpdateIndustryDivision() {
  return useUpdateMutation<IndustryDivisionPayload, IndustryDivision>("industry-divisions", INDUSTRYDIVISIONS_URL);
}

export function useDeleteIndustryDivision() {
  return useHierarchyDeleteMutation("industry-divisions", INDUSTRYDIVISIONS_URL, INDUSTRY_QUERY_KEYS);
}


// INDUSTRYGROUP ("industry-groups") - industry hierarchy level

const INDUSTRYGROUPS_URL = `${BASE}/industry/industry-groups`;

export function useIndustryGroups() {
  return useListQuery<IndustryGroupListResponse>("industry-groups", INDUSTRYGROUPS_URL);
}

export function useCreateIndustryGroup() {
  return useCreateMutation<IndustryGroupPayload, IndustryGroup>("industry-groups", INDUSTRYGROUPS_URL);
}

export function useUpdateIndustryGroup() {
  return useUpdateMutation<IndustryGroupPayload, IndustryGroup>("industry-groups", INDUSTRYGROUPS_URL);
}

export function useDeleteIndustryGroup() {
  return useHierarchyDeleteMutation("industry-groups", INDUSTRYGROUPS_URL, INDUSTRY_QUERY_KEYS);
}


// INDUSTRYCLASS ("industry-classes") - industry hierarchy level

const INDUSTRYCLASSES_URL = `${BASE}/industry/industry-classes`;

export function useIndustryClasses() {
  return useListQuery<IndustryClassListResponse>("industry-classes", INDUSTRYCLASSES_URL);
}

export function useCreateIndustryClass() {
  return useCreateMutation<IndustryClassPayload, IndustryClass>("industry-classes", INDUSTRYCLASSES_URL);
}

export function useUpdateIndustryClass() {
  return useUpdateMutation<IndustryClassPayload, IndustryClass>("industry-classes", INDUSTRYCLASSES_URL);
}

export function useDeleteIndustryClass() {
  return useHierarchyDeleteMutation("industry-classes", INDUSTRYCLASSES_URL, INDUSTRY_QUERY_KEYS);
}


// INDUSTRYSUBCLASS ("industry-subclasses") - industry hierarchy level

const INDUSTRYSUBCLASSES_URL = `${BASE}/industry/industry-subclasses`;

export function useIndustrySubclasses(limit?: number, offset?: number) {
   return useListQuery<IndustrySubclassListResponse>("industry-subclasses", INDUSTRYSUBCLASSES_URL, { limit, offset });
}

export function useCreateIndustrySubclass() {
  return useCreateMutation<IndustrySubclassPayload, IndustrySubclass>("industry-subclasses", INDUSTRYSUBCLASSES_URL);
}

export function useUpdateIndustrySubclass() {
  return useUpdateMutation<IndustrySubclassPayload, IndustrySubclass>("industry-subclasses", INDUSTRYSUBCLASSES_URL);
}

export function useDeleteIndustrySubclass() {
  return useHierarchyDeleteMutation("industry-subclasses", INDUSTRYSUBCLASSES_URL, INDUSTRY_QUERY_KEYS);
}
