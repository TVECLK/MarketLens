function InfoIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M12 16v-4" />
      <path d="M12 8h.01" />
    </svg>
  );
}

/** Mobile-only tip (hidden from `md` up) shown above the charts. */
export default function RotateDeviceHint() {
  return (
    <div className="md:hidden mx-4 mt-4 bg-indigo-50 border border-indigo-200 text-indigo-700 text-xs font-medium px-4 py-2 rounded-xl flex items-center gap-2">
      <InfoIcon className="size-4 shrink-0" />
      Rotate your phone to landscape for a better view of the charts
    </div>
  );
}
