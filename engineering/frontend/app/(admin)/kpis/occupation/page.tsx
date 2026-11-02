import { redirect } from "next/navigation";
import { hierarchyLevelHref } from "@/lib/kpi-config";

export default function OccupationIndexPage() {
  redirect(hierarchyLevelHref("occupation", 0));
}
