import { redirect } from "next/navigation";
import { CATEGORY_ORDER, kpiCategoryHref } from "@/lib/kpi-config";

export default function KpisIndexPage() {
  redirect(kpiCategoryHref(CATEGORY_ORDER[0]));
}
