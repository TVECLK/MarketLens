import { notFound } from "next/navigation";
import { categoryKeyFromSlug } from "@/lib/kpi-config";
import FlatCategoryRoute from "@/components/kpis/FlatCategoryRoute";

export default async function KpiCategoryPage({
  params,
}: {
  params: Promise<{ category: string }>;
}) {
  const { category } = await params;
  const key = categoryKeyFromSlug(category);
  if (!key) notFound();

  return <FlatCategoryRoute category={key} />;
}
