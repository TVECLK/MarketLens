import type { ReactNode } from "react";
import { clsx } from "clsx";

interface ChartCardProps {
  title: ReactNode;
  subtitle?: string;
  heightClassName: string;
  bodyClassName?: string;
  headerRight?: ReactNode;
  footer?: ReactNode;
  isLoading: boolean;
  loadingLabel: string;
  isError?: boolean;
  isEmpty?: boolean;
  variant?: "default" | "public";
  children: ReactNode;
}

function CardMessage({ children }: { children: ReactNode }) {
  return (
    <div className="w-full h-full flex items-center justify-center text-xs text-zinc-400">
      {children}
    </div>
  );
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
  isError = false,
  isEmpty = false,
  variant = "default",
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
        <div className={variant === "public" ? "flex flex-col mb-2" : undefined}>
          <h4
            className={
              variant === "public"
                ? "w-2/3 text-2xl font-normal text-black-600 whitespace-pre-line leading-snug"
                : "text-sm font-bold text-zinc-900"
            }
          >
            {title}
          </h4>
          {subtitle && (
            variant === "public"
              ? <p className="text-xl text-zinc-400 mt-auto pt-8">{subtitle}</p>
              : <p className="text-[11px] text-zinc-400">{subtitle}</p>
          )}
        </div>
        {headerRight}
      </div>
      <div className={clsx("flex-1 mt-4 min-w-0", bodyClassName)}>
        {isLoading ? (
          <CardMessage>{loadingLabel}</CardMessage>
        ) : isError ? (
          <CardMessage>Couldn&apos;t load data</CardMessage>
        ) : isEmpty ? (
          <CardMessage>No data for this range</CardMessage>
        ) : (
          children
        )}
      </div>
      {footer}
    </div>
  );
}
