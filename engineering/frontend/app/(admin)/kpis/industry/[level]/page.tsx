import { notFound } from "next/navigation";
import { levelIndexFromSlug } from "@/lib/kpi-config";
import HierarchyRoute from "@/components/kpis/HierarchyRoute";

export default async function IndustryLevelPage({
  params,
}: {
  params: Promise<{ level: string }>;
}) {
  const { level } = await params;
  const levelIndex = levelIndexFromSlug("industry", level);
  if (levelIndex === -1) notFound();

  return <HierarchyRoute hierarchyId="industry" levelIndex={levelIndex} />;
}
