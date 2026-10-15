import {
  KPI_CATEGORIES,
  CATEGORY_ORDER,
  HIERARCHIES,
  HIERARCHY_ORDER,
  ACCENT_CLASSES,
  type KPICategoryKey,
  type HierarchyId,
} from "@/lib/kpi-config";
import type { SectionKey } from "@/types/kpi-admin";

interface CategoryNavProps {
  activeSection: SectionKey;
  activeCategory: KPICategoryKey;
  flatCounts: Record<KPICategoryKey, number>;
  onSwitchCategory: (key: KPICategoryKey) => void;
  onSwitchHierarchy: (id: HierarchyId) => void;
}

export default function CategoryNav({
  activeSection,
  activeCategory,
  flatCounts,
  onSwitchCategory,
  onSwitchHierarchy,
}: CategoryNavProps) {
  return (
    <aside className="w-72 min-w-72 border-r border-gray-100 bg-white h-[calc(100vh-89px)] overflow-y-auto sticky top-[89px] py-4 px-3 space-y-1">
      {CATEGORY_ORDER.map((key) => {
        const cfg = KPI_CATEGORIES[key];
        const acc = ACCENT_CLASSES[cfg.accent];
        const count = flatCounts[key];
        const isActive = activeSection === "flat" && activeCategory === key;
        return (
          <button
            key={key}
            onClick={() => onSwitchCategory(key)}
            className={`w-full text-left px-4 py-3 rounded-xl transition-all flex items-center justify-between gap-2 ${
              isActive ? `${acc.lightBg} ring-1 ${acc.ring}` : "hover:bg-gray-50"
            }`}
          >
            <div>
              <p className={`text-xs font-bold ${isActive ? acc.text : "text-gray-700"}`}>{cfg.title}</p>
              <p className="text-[10px] text-gray-400 mt-0.5">{count} items</p>
            </div>
            <span className={`w-2 h-2 rounded-full shrink-0 ${isActive ? acc.solidBg : "bg-gray-200"}`} />
          </button>
        );
      })}

      <div className="pt-3 mt-2 border-t border-gray-100">
        <p className="px-4 pb-1.5 text-[10px] font-black uppercase text-gray-300 tracking-wider">
          Hierarchical
        </p>
        {HIERARCHY_ORDER.map((id) => {
          const def = HIERARCHIES[id];
          const acc = ACCENT_CLASSES[def.accent];
          const isActive = activeSection === id;
          return (
            <button
              key={id}
              onClick={() => onSwitchHierarchy(id)}
              className={`w-full text-left px-4 py-3 rounded-xl transition-all flex items-center justify-between gap-2 ${
                isActive ? `${acc.lightBg} ring-1 ${acc.ring}` : "hover:bg-gray-50"
              }`}
            >
              <div>
                <p className={`text-xs font-bold ${isActive ? acc.text : "text-gray-700"}`}>{def.label}</p>
                <p className="text-[10px] text-gray-400 mt-0.5">{def.levels.length} levels</p>
              </div>
              <span className={`w-2 h-2 rounded-full shrink-0 ${isActive ? acc.solidBg : "bg-gray-200"}`} />
            </button>
          );
        })}
      </div>
    </aside>
  );
}
