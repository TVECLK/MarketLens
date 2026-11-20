import { Suspense } from "react";
import Header from "@/components/layout/Header";
import LoadingSpinner from "@/components/shared/LoadingSpinner";
import RotateDeviceHint from "@/components/shared/RotateDeviceHint";
import DashboardDateRangePicker from "@/components/dashboard/DashboardDateRangePicker";
import { DateRangeProvider } from "@/providers/date-range-provider";
import logo from "@/public/logo.png";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen w-full min-w-0 bg-zinc-50 text-zinc-800 font-sans">
      <Header
        title="Labour Market Demand Dashboard"
        subtitle="National overview of labour market demand across occupations and industries"
        logo={logo}
      />
      <Suspense
        fallback={
          <LoadingSpinner label="Loading..." className="py-20 text-xs text-zinc-400" />
        }
      >
        <DateRangeProvider>
          <DashboardDateRangePicker />
          <RotateDeviceHint />
          {children}
        </DateRangeProvider>
      </Suspense>
    </div>
  );
}
