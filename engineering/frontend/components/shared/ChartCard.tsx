import type { ReactNode } from "react";
import { clsx } from "clsx";

interface ChartCardProps {
  title: ReactNode;
  subtitle?: string;
  headline?: ReactNode;
  /** Show a two-line placeholder in the headline's place while loading (for cards that have one). */
  headlineSkeleton?: boolean;
  heightClassName: string;
  bodyClassName?: string;
  headerRight?: ReactNode;
  footer?: ReactNode;
  isLoading: boolean;
  loadingLabel: string;
  isError?: boolean;
  isEmpty?: boolean;
  children: ReactNode;
}

/** Two pulsing bars sized like the headline text, matching the stat cards' skeleton. */
function HeadlineSkeleton() {
  return (
    <div aria-hidden className="w-full md:w-2/3 flex flex-col gap-2 mb-2">
      <div className="h-5 md:h-7 w-full bg-zinc-100 rounded-md animate-pulse" />
      <div className="h-5 md:h-7 w-2/3 bg-zinc-100 rounded-md animate-pulse" />
    </div>
  );
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
  headline,
  headlineSkeleton = false,
  heightClassName,
  bodyClassName,
  headerRight,
  footer,
  isLoading,
  loadingLabel,
  isError = false,
  isEmpty = false,
  children,
}: ChartCardProps) {
  return (
    <div
      className={clsx(
        "bg-white border border-zinc-200 p-5 rounded-xl shadow-sm flex flex-col overflow-hidden",
        heightClassName,
      )}
    >
      <div className="flex flex-col gap-2">
        {headlineSkeleton && isLoading && (
          <>
            <HeadlineSkeleton />
            <div className="border-b border-zinc-100" />
          </>
        )}
        {headline && !isLoading && !isError && (
          <>
            <p className="w-full md:w-2/3 text-md md:text-xl md:text-2xl font-normal text-zinc-900 whitespace-pre-line leading-snug mb-2">{headline}</p>
            <div className="border-b border-zinc-100" />
          </>
        )}
        <div className={clsx("flex flex-wrap items-start gap-2", headerRight && "justify-between")}>
          <div>
            <h4 className="text-base font-semibold text-zinc-900 leading-none">{title}</h4>
            {subtitle && <p className="text-sm text-zinc-600 leading-none mt-2">{subtitle}</p>}
          </div>
          {headerRight}
        </div>
        <div className="border-b border-zinc-100" />
      </div>
      <div className={clsx("flex-1 mt-2 min-w-0 min-h-0", bodyClassName)}>
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
