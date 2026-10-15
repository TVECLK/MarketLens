import type { ReactNode } from "react";
import { clsx } from "clsx";

interface ChartCardProps {
  title: string;
  subtitle: string;
  heightClassName: string;
  bodyClassName?: string;
  headerRight?: ReactNode;
  footer?: ReactNode;
  isLoading: boolean;
  loadingLabel: string;
  children: ReactNode;
}

export default function ChartCard({
  title,
  subtitle,
  heightClassName,
  bodyClassName,
  headerRight,
  footer,
  isLoading,
  loadingLabel,
  children,
}: ChartCardProps) {
  return (
    <div
      className={clsx(
        "bg-white border border-zinc-200 p-5 rounded-xl shadow-sm flex flex-col",
        heightClassName,
      )}
    >
      <div
        className={clsx(
          "flex flex-wrap items-start border-b border-zinc-100 pb-2 gap-2",
          headerRight && "justify-between",
        )}
      >
        <div>
          <h4 className="text-sm font-bold text-zinc-900">{title}</h4>
          <p className="text-[11px] text-zinc-400">{subtitle}</p>
        </div>
        {headerRight}
      </div>
      <div className={clsx("flex-1 mt-4 min-w-0", bodyClassName)}>
        {isLoading ? (
          <div className="w-full h-full flex items-center justify-center text-xs text-zinc-400">
            {loadingLabel}
          </div>
        ) : (
          children
        )}
      </div>
      {footer}
    </div>
  );
}
