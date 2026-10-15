import DataTable from "@/components/shared/DataTable";
import type { ChartDatum } from "@/types/chart";

interface AnalysisBreakdownTablesProps {
  childrenBreakdown: ChartDatum[];
  childrenBreakdownTitle: string;
  sectorData: ChartDatum[];
  expData: ChartDatum[];
  provinceData: ChartDatum[];
  eduData: ChartDatum[];
  formalData: ChartDatum[];
  genderData: ChartDatum[];
  nvqData: ChartDatum[];
  remoteData: ChartDatum[];
  contractData: ChartDatum[];
}

export default function AnalysisBreakdownTables({
  childrenBreakdown,
  childrenBreakdownTitle,
  sectorData,
  expData,
  provinceData,
  eduData,
  formalData,
  genderData,
  nvqData,
  remoteData,
  contractData,
}: AnalysisBreakdownTablesProps) {
  return (
    <div className="p-4 md:p-5 grid grid-cols-1 xl:grid-cols-2 gap-6">
      {childrenBreakdown.length > 0 && (
        <DataTable title={childrenBreakdownTitle} rows={childrenBreakdown} />
      )}
      <DataTable title="Employment Sector" rows={sectorData} />
      <DataTable title="Experience" rows={expData} />
      <DataTable title="Province wise Distribution" rows={provinceData} />
      <DataTable title="Education Level" rows={eduData} />
      <DataTable title="Formal / Informal Sector" rows={formalData} />
      <DataTable title="Gender" rows={genderData} />
      <DataTable title="Vocational Education (NVQ Level)" rows={nvqData} />
      <DataTable title="Remote / On-Site" rows={remoteData} />
      <DataTable title="Contract Type" rows={contractData} />
    </div>
  );
}
