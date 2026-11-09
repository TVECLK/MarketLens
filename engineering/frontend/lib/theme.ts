// THEME — the one colour scheme of the app chrome and the KPI admin screens.
// Tailwind needs literal class names, so each entry is a ready-made class
// string; change a colour here and every screen follows.
export const THEME = {
  /** The dark-blue main sidebar (shown when signed in). */
  sidebar: {
    bg: "bg-[#1B365D]",
    border: "border-white/10",
    subtext: "text-blue-200",
    toggle: "bg-[#1B365D] border-white/20 text-blue-200 hover:text-white hover:border-white/40",
    link: "text-blue-100 hover:bg-white/10 hover:text-white",
    linkActive: "bg-white text-[#1B365D] font-semibold shadow-sm",
  },
  /** The fill for every action button. */
  button: "bg-[#005A9C] text-white hover:bg-[#004a80]",
  /** The grey fill for secondary buttons (Cancel, Reset). */
  buttonSecondary: "bg-gray-100 text-gray-500 hover:bg-gray-200",
  /** Form controls and their labels. */
  form: {
    input:
      "w-full bg-gray-50 rounded-lg p-2.5 text-xs font-semibold border border-gray-200 outline-none focus:ring-2 focus:ring-offset-0",
    label: "text-[10px] font-black uppercase text-gray-400 tracking-wider block mb-1.5",
  },
  /** The KPI category list. */
  nav: {
    item: "hover:bg-gray-50",
    itemActive: "bg-blue-50 ring-1 ring-blue-200",
    title: "text-gray-700",
    titleActive: "text-blue-700",
  },
  /** The selected row of a tree. */
  selectedRow: "bg-blue-50",
  selectedText: "text-blue-700",
  /** Small level label. */
  chip: "bg-blue-50 text-blue-700",
} as const;
