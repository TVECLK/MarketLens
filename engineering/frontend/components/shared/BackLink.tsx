import Link from "next/link";

interface BackLinkProps {
  href: string;
}

export default function BackLink({ href }: BackLinkProps) {
  return (
    <Link
      href={href}
      className="cursor-pointer inline-flex items-center gap-1.5 text-xs font-bold text-zinc-500 hover:text-zinc-900 transition-colors"
    >
      <svg
        className="w-3.5 h-3.5"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <polyline points="15 18 9 12 15 6" />
      </svg>
      Back
    </Link>
  );
}
