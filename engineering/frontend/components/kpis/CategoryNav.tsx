"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  useEmploymentSectors,
  useExperiences,
  useEducationLevels,
  useFormalities,
  useGenders,
  useVocationalEducations,
} from "@/hooks/use-kpi";
import {
  KPI_CATEGORIES,
  CATEGORY_ORDER,
  HIERARCHIES,
  HIERARCHY_ORDER,
  ACCENT_CLASSES,
  kpiCategoryHref,
  hierarchyLevelHref,
  type KPICategoryKey,
} from "@/lib/kpi-config";

// access the values from the cache
function useFlatCounts(): Record<KPICategoryKey, number | undefined> {
  const employmentSectors = useEmploymentSectors();
  const experiences = useExperiences();
  const educationLevels = useEducationLevels();
  const formalities = useFormalities();
  const genders = useGenders();
  const vocationalEducations = useVocationalEducations();
  return {
    employmentSector: employmentSectors.data?.employment_sectors?.length,
    experience: experiences.data?.experiences?.length,
    educationLevel: educationLevels.data?.education_levels?.length,
    formalInformal: formalities.data?.formalities?.length,
    gender: genders.data?.genders?.length,
    vocationalEducation: vocationalEducations.data?.vocational_educations?.length,
  };
}

export default function CategoryNav() {
  const pathname = usePathname();
  const flatCounts = useFlatCounts();

  return (
    <aside className="border-r border-gray-100 bg-white h-screen overflow-y-auto sticky top-0 p-3 space-y-1">
      {CATEGORY_ORDER.map((key) => {
        const cfg = KPI_CATEGORIES[key];
        const acc = ACCENT_CLASSES[cfg.accent];
        const count = flatCounts[key];
        const href = kpiCategoryHref(key);
        const isActive = pathname === href;
        return (
          <Link
            key={key}
            href={href}
            aria-current={isActive ? "page" : undefined}
            className={`w-full text-left px-3 py-2 rounded-md transition-all flex items-center justify-between gap-2 ${isActive ? `${acc.lightBg} ring-1 ${acc.ring}` : "hover:bg-gray-50"
              }`}
          >
            <div>
              <p className={`text-xs font-bold ${isActive ? acc.text : "text-gray-700"}`}>{cfg.title}</p>
              <p className="text-[10px] text-gray-400 mt-0.5">{count ?? "–"} items</p>
            </div>
          </Link>
        );
      })}

      <div className="pt-3 mt-2 border-t border-gray-100">
        <p className="px-4 pb-1.5 text-[10px] font-black uppercase text-gray-300 tracking-wider">
          Hierarchical
        </p>
        <div className="space-y-1">
          {HIERARCHY_ORDER.map((id) => {
            const def = HIERARCHIES[id];
            const acc = ACCENT_CLASSES[def.accent];
            const isActive = pathname.startsWith(`/kpis/${id}/`);
            return (
              <Link
                key={id}
                href={hierarchyLevelHref(id, 0)}
                aria-current={isActive ? "page" : undefined}
                className={`w-full text-left px-3 py-2 rounded-md transition-all flex items-center justify-between gap-2 ${isActive ? `${acc.lightBg} ring-1 ${acc.ring}` : "hover:bg-gray-50"
                  }`}
              >
                <div>
                  <p className={`text-xs font-bold ${isActive ? acc.text : "text-gray-700"}`}>{def.label}</p>
                  <p className="text-[10px] text-gray-400 mt-0.5">{def.levels.length} levels</p>
                </div>
              </Link>
            );
          })}
        </div>
      </div>
    </aside>
  );
}
