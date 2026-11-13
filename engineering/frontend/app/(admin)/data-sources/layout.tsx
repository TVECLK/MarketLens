import Header from "@/components/layout/Header";
import PermissionsProvider from "@/components/kpis/PermissionsProvider";

export default function DataSourcesLayout({ children }: { children: React.ReactNode }) {
  return (
    <PermissionsProvider>
      <div className="min-h-screen bg-gray-50 text-gray-800 font-sans">
        <Header
          title="Data Sources"
          subtitle="Turn job sources on or off, and manage hand-entered sources such as newspapers"
        />

        <div className="flex">{children}</div>
      </div>
    </PermissionsProvider>
  );
}
