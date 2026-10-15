import type { ChartDatum } from "@/types/chart";

interface DataTableProps {
  title: string;
  rows: ChartDatum[];
}

export default function DataTable({ title, rows }: DataTableProps) {
  const sorted = [...rows].sort((a, b) => b.value - a.value);
  return (
    <div className="border border-zinc-100 p-5 rounded-xl flex flex-col">
      <h4 className="text-sm font-bold text-zinc-900 border-b border-zinc-100 pb-2">
        {title}
      </h4>
      <div className="mt-3 overflow-x-auto rounded-lg border border-zinc-100">
        <table className="w-full text-left">
          <thead className="bg-zinc-50">
            <tr className="border-b border-zinc-200">
              <th className="px-3 py-2 text-[10px] font-black uppercase tracking-wider text-zinc-400 w-10">
                #
              </th>
              <th className="px-3 py-2 text-[10px] font-black uppercase tracking-wider text-zinc-400">
                Name
              </th>
              <th className="px-3 py-2 text-[10px] font-black uppercase tracking-wider text-zinc-400 text-right">
                Job Count
              </th>
            </tr>
          </thead>
          <tbody>
            {sorted.length === 0 ? (
              <tr>
                <td
                  colSpan={3}
                  className="px-3 py-6 text-center text-xs text-zinc-400"
                >
                  No data available.
                </td>
              </tr>
            ) : (
              sorted.map((row, i) => (
                <tr
                  key={row.name}
                  className="border-b border-zinc-50 last:border-0 hover:bg-zinc-50 transition-colors"
                >
                  <td className="px-3 py-2 text-xs font-bold text-zinc-400">
                    {i + 1}
                  </td>
                  <td className="px-3 py-2 text-xs font-bold text-zinc-800">
                    {row.name}
                  </td>
                  <td className="px-3 py-2 text-xs font-mono font-bold text-zinc-800 text-right">
                    {row.value.toLocaleString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
