export interface EmploymentSector {
  id: number;
  sector: string;
  created_at: string;
  updated_at: string;
}

export interface Experience {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface EducationLevel {
  id: number;
  level: string;
  created_at: string;
  updated_at: string;
}

export interface Formality {
  id: number;
  formality_type: string;
  created_at: string;
  updated_at: string;
}

export interface Gender {
  id: number;
  gender_type: string;
  created_at: string;
  updated_at: string;
}

export interface VocationalEducation {
  id: number;
  level: string;
  created_at: string;
  updated_at: string;
}

// List response wrappers (GET all)

export interface EmploymentSectorListResponse {
  count: number;
  employment_sectors: EmploymentSector[];
}
export interface ExperienceListResponse {
  count: number;
  experiences: Experience[];
}
export interface EducationLevelListResponse {
  count: number;
  education_levels: EducationLevel[];
}
export interface FormalityListResponse {
  count: number;
  formalities: Formality[];
}
export interface GenderListResponse {
  count: number;
  genders: Gender[];
}
export interface VocationalEducationListResponse {
  count: number;
  vocational_educations: VocationalEducation[];
}

// Create/update payloads
// PUT endpoints accept a partial "updates" map, so all fields are optional.

export interface EmploymentSectorPayload {
  sector: string;
}
export interface ExperiencePayload {
  name: string;
}
export interface EducationLevelPayload {
  level: string;
}
export interface FormalityPayload {
  formality_type: string;
}
export interface GenderPayload {
  gender_type: string;
}
export interface VocationalEducationPayload {
  level: string;
}


// OCCUPATION HIERARCHY
// Major Group -> Sub Major Group -> Minor Group -> Unit Group -> Occupation Group

export interface MajorGroup {
  id: number;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface SubMajorGroup {
  id: number;
  major_group_id: number;
  major_group: MajorGroup | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface MinorGroup {
  id: number;
  sub_major_group_id: number;
  sub_major_group: SubMajorGroup | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface UnitGroup {
  id: number;
  minor_group_id: number;
  minor_group: MinorGroup | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface OccupationGroup {
  id: number;
  unit_group_id: number;
  unit_group: UnitGroup | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface MajorGroupListResponse {
  count: number;
  major_groups: MajorGroup[];
}

export interface MajorGroupPayload {
  name: string;
  code: string;
}
export interface SubMajorGroupPayload {
  name: string;
  code: string;
  major_group_id: number;
}
export interface MinorGroupPayload {
  name: string;
  code: string;
  sub_major_group_id: number;
}
export interface UnitGroupPayload {
  name: string;
  code: string;
  minor_group_id: number;
}
export interface OccupationGroupPayload {
  name: string;
  code: string;
  unit_group_id: number;
}


// INDUSTRY HIERARCHY
// Industry Sector -> Industry Division -> Industry Group -> Industry Class
// -> Industry Subclass

export interface IndustrySector {
  id: number;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface IndustryDivision {
  id: number;
  industry_sector_id: number;
  industry_sector: IndustrySector | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface IndustryGroup {
  id: number;
  industry_division_id: number;
  industry_division: IndustryDivision | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface IndustryClass {
  id: number;
  industry_group_id: number;
  industry_group: IndustryGroup | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface IndustrySubclass {
  id: number;
  industry_class_id: number;
  industry_class: IndustryClass | null;
  name: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface IndustrySectorListResponse {
  count: number;
  industry_sectors: IndustrySector[];
}

export interface IndustrySectorPayload {
  name: string;
  code: string;
}
export interface IndustryDivisionPayload {
  name: string;
  code: string;
  industry_sector_id: number;
}
export interface IndustryGroupPayload {
  name: string;
  code: string;
  industry_division_id: number;
}
export interface IndustryClassPayload {
  name: string;
  code: string;
  industry_group_id: number;
}
export interface IndustrySubclassPayload {
  name: string;
  code: string;
  industry_class_id: number;
}


// GENERIC HELPERS - a common shape every hierarchy level's UI can render
// against, regardless of which specific type/payload it wraps.

/** Normalized shape the UI works with for any hierarchy level's row. */
export interface HierarchyRow {
  id: number;
  name: string;
  code: string;
  parentId: number | null;
  parentLabel: string | null; // e.g. "11 — Chief Executives..." or null at top level
}
