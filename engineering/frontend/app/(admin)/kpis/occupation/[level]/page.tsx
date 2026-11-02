import { notFound } from "next/navigation";
import { levelIndexFromSlug } from "@/lib/kpi-config";
import HierarchyRoute from "@/components/kpis/HierarchyRoute";

export default async function OccupationLevelPage({
  params,
}: {
  params: Promise<{ level: string }>;
}) {
  const { level } = await params;
  const levelIndex = levelIndexFromSlug("occupation", level);
  if (levelIndex === -1) notFound();

  return <HierarchyRoute hierarchyId="occupation" levelIndex={levelIndex} />;
}
