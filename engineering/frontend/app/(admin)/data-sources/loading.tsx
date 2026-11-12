export default function DataSourcesLoading() {
  return (
    <div className="flex-1 h-screen flex items-center justify-center bg-gray-50" role="status" aria-live="polite">
      <div className="flex flex-col items-center gap-3">
        <div className="w-8 h-8 rounded-full border-2 border-gray-200 border-t-gray-800 animate-spin" />
        <p className="text-xs font-semibold text-gray-500">Checking your access…</p>
      </div>
    </div>
  );
}
