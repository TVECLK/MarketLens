import DataTable from "@/components/shared/DataTable";
import type { AnalysisSectionErrors } from "@/types/analysis";
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
  sectionErrors: AnalysisSectionErrors;
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
  sectionErrors,
}: AnalysisBreakdownTablesProps) {
  return (
    <div className="p-4 md:p-5 grid grid-cols-1 xl:grid-cols-2 gap-6">
      {childrenBreakdown.length > 0 && (
        <DataTable title={childrenBreakdownTitle} rows={childrenBreakdown} />
      )}
      <DataTable
        title="Employment Sector"
        rows={sectorData}
        isError={Boolean(sectionErrors.employment_sector)}
      />
      <DataTable
        title="Experience"
        rows={expData}
        isError={Boolean(sectionErrors.experience)}
      />
      <DataTable
        title="Province wise Distribution"
        rows={provinceData}
        isError={Boolean(sectionErrors.province)}
      />
      <DataTable
        title="Education Level"
        rows={eduData}
        isError={Boolean(sectionErrors.education)}
      />
      <DataTable
        title="Formal / Informal Sector"
        rows={formalData}
        isError={Boolean(sectionErrors.formality)}
      />
      <DataTable
        title="Gender"
        rows={genderData}
        isError={Boolean(sectionErrors.gender)}
      />
      <DataTable
        title="Vocational Education (NVQ Level)"
        rows={nvqData}
        isError={Boolean(sectionErrors.vocational_education)}
      />
      <DataTable
        title="Remote / On-Site"
        rows={remoteData}
        isError={Boolean(sectionErrors.remote_onsite_hybrid)}
      />
      <DataTable
        title="Contract Type"
        rows={contractData}
        isError={Boolean(sectionErrors.job_type)}
      />
    </div>
  );
}
