import DataTable from "@/components/shared/DataTable";
import { BREAKDOWN_SECTIONS, CHILDREN_BREAKDOWN_TOOLTIP } from "@/components/analysis/breakdown-sections";
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
        <DataTable title={childrenBreakdownTitle} tooltip={CHILDREN_BREAKDOWN_TOOLTIP} rows={childrenBreakdown} />
      )}
      <DataTable
        {...BREAKDOWN_SECTIONS.employment_sector}
        rows={sectorData}
        isError={Boolean(sectionErrors.employment_sector)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.experience}
        rows={expData}
        isError={Boolean(sectionErrors.experience)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.province}
        rows={provinceData}
        isError={Boolean(sectionErrors.province)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.education}
        rows={eduData}
        isError={Boolean(sectionErrors.education)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.formality}
        rows={formalData}
        isError={Boolean(sectionErrors.formality)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.gender}
        rows={genderData}
        isError={Boolean(sectionErrors.gender)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.vocational_education}
        rows={nvqData}
        isError={Boolean(sectionErrors.vocational_education)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.work_modes}
        rows={remoteData}
        isError={Boolean(sectionErrors.work_modes)}
      />
      <DataTable
        {...BREAKDOWN_SECTIONS.job_type}
        rows={contractData}
        isError={Boolean(sectionErrors.job_type)}
      />
    </div>
  );
}
