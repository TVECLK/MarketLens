import { InformationCircleIcon } from "@heroicons/react/24/outline";

/** Mobile-only tip (hidden from `md` up) shown above the charts. */
export default function RotateDeviceHint() {
  return (
    <div className="md:hidden mx-4 mt-4 bg-indigo-50 border border-indigo-200 text-indigo-700 text-xs font-medium px-4 py-2 rounded-xl flex items-center gap-2">
      <InformationCircleIcon className="size-4 shrink-0" />
      Rotate your phone to landscape for a better view of the charts
    </div>
  );
}
