"use client";

export type AnalysisViewMode = "chart" | "table";

interface ViewModeToggleProps {
  viewMode: AnalysisViewMode;
  onChange: (mode: AnalysisViewMode) => void;
}

export default function ViewModeToggle({ viewMode, onChange }: ViewModeToggleProps) {
  return (
    <div className="px-5 py-3 border-b border-zinc-200 flex items-center gap-3">
      <div className="flex bg-zinc-100 rounded-lg p-0.5">
        <button
          onClick={() => onChange("chart")}
          className={`cursor-pointer px-3 py-1.5 text-xs font-bold rounded-md transition-colors ${viewMode === "chart" ? "bg-white text-zinc-900 shadow-sm" : "text-zinc-500 hover:text-zinc-700"}`}
        >
          Chart View
        </button>
        <button
          onClick={() => onChange("table")}
          className={`cursor-pointer px-3 py-1.5 text-xs font-bold rounded-md transition-colors ${viewMode === "table" ? "bg-white text-zinc-900 shadow-sm" : "text-zinc-500 hover:text-zinc-700"}`}
        >
          Table View
        </button>
      </div>
    </div>
  );
}
