import { redirect } from "next/navigation";
import { hierarchyLevelHref } from "@/lib/kpi-config";

export default function IndustryIndexPage() {
  redirect(hierarchyLevelHref("industry", 0));
}
