import Link from "next/link";
import { ChevronLeftIcon } from "@heroicons/react/16/solid";

interface BackLinkProps {
  href: string;
}

export default function BackLink({ href }: BackLinkProps) {
  return (
    <Link
      href={href}
      className="cursor-pointer inline-flex items-center gap-1.5 text-xs font-bold text-zinc-500 hover:text-zinc-900 transition-colors"
    >
      <ChevronLeftIcon className="w-3.5 h-3.5" />
      Back
    </Link>
  );
}
