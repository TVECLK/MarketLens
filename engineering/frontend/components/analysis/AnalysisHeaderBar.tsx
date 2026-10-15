interface AnalysisHeaderBarProps {
  levelLabel: string;
  nodeName: string;
  nodeTotal: number;
  isLoading: boolean;
}

export default function AnalysisHeaderBar({
  levelLabel,
  nodeName,
  nodeTotal,
  isLoading,
}: AnalysisHeaderBarProps) {
  return (
    <div className="px-5 py-4 border-b border-zinc-200 bg-zinc-50/70 rounded-t-xl flex flex-wrap justify-between items-end gap-4">
      <div className="min-w-0">
        <p className="text-[10px] font-black uppercase tracking-wider text-zinc-400 mb-1">
          Analytics for {levelLabel}
        </p>
        <p className="text-base font-black text-zinc-900 truncate">{nodeName}</p>
      </div>
      <div className="text-right shrink-0">
        <p className="text-[10px] font-black uppercase tracking-wider text-zinc-400">
          Vacancies for selection
        </p>
        <p className="text-lg font-black text-zinc-900">
          {isLoading ? "…" : nodeTotal.toLocaleString()}
        </p>
      </div>
    </div>
  );
}
