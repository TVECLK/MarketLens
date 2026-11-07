import Header from "@/components/layout/Header";
import CategoryNav from "@/components/kpis/CategoryNav";
import PermissionsProvider from "@/components/kpis/PermissionsProvider";

export default function KpisLayout({ children }: { children: React.ReactNode }) {
  return (
    <PermissionsProvider>
      <div className="min-h-screen bg-gray-50 text-gray-800 font-sans">
        <Header
          title="KPI Reference Data Manager"
          subtitle="Create, edit, and remove the reference values that power the labour market dashboard charts"
        />

        <div className="flex">
          <CategoryNav />
          {children}
        </div>
      </div>
    </PermissionsProvider>
  );
}
