"use client";

import { useId, useState } from "react";
import { LockClosedIcon } from "@heroicons/react/16/solid";
import SourceDialog from "@/components/sources/SourceDialog";
import { getErrorMessage } from "@/lib/errors";
import {
  findCaseOnlyNameMatch,
  findExactNameMatch,
  hasInnerWhitespace,
  hasOuterWhitespace,
  normalizeSourceName,
  visibleWhitespace,
} from "@/lib/source-name";
import { THEME } from "@/lib/theme";
import { SOURCE_TYPE_LABELS, type CreateSourcePayload, type Source, type SourceType } from "@/types/source";

interface AddSourceModalProps {
  existing: readonly Source[];
  create: (payload: CreateSourcePayload, callbacks: { onSuccess: (source: Source) => void; onError: (err: Error) => void }) => void;
  isPending: boolean;
  onCancel: () => void;
  onCreated: (source: Source) => void;
}

const TYPE_CARDS: { type: SourceType; title: string; description: string }[] = [
  { type: "auto", title: SOURCE_TYPE_LABELS.auto, description: "Collected automatically from a job website. The name must match exactly with the crawler registry name" },
  { type: "manual", title: SOURCE_TYPE_LABELS.manual, description: "Entered by hand, e.g. newspaper ads. Can be renamed." },
];

/**
 * Adds a source. One form that adapts to the type: an auto source's name must exactly match its
 * crawler's SOURCE_NAME, so for auto the name is shown exactly as it will be saved, mistakes that
 * break the match are flagged, and the admin confirms before adding.
 */
export default function AddSourceModal({ existing, create, isPending, onCancel, onCreated }: AddSourceModalProps) {
  const [type, setType] = useState<SourceType>("manual");
  const [name, setName] = useState("");
  const [isActive, setIsActive] = useState(true);
  const [confirmed, setConfirmed] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const nameId = useId();
  const confirmId = useId();

  const isAuto = type === "auto";
  const saved = normalizeSourceName(name);
  const taken = findExactNameMatch(name, existing);
  const caseOnlyMatch = findCaseOnlyNameMatch(name, existing);
  const canSubmit = !isPending && saved !== "" && !taken && (!isAuto || confirmed);

  // A confirmation is for one exact name and type: changing either asks again.
  function chooseType(next: SourceType) {
    setType(next);
    setConfirmed(false);
    setError(null);
  }
  function changeName(next: string) {
    setName(next);
    setConfirmed(false);
    setError(null);
  }

  function submit() {
    if (!canSubmit) return;
    create(
      { name: saved, type, is_active: isActive },
      {
        onSuccess: onCreated,
        onError: (err) => setError(getErrorMessage(err, "Failed to add source.")),
      },
    );
  }

  return (
    <SourceDialog
      title="Add source"
      onClose={onCancel}
      onSubmit={submit}
      size="md"
      footer={
        <>
          <button
            type="button"
            onClick={onCancel}
            className={`cursor-pointer text-xs font-bold ${THEME.buttonSecondary} px-4 py-2 rounded-full transition-all`}
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={!canSubmit}
            title={isAuto && !confirmed && saved ? "Confirm the name matches the crawler first" : undefined}
            className={`cursor-pointer text-xs font-bold ${THEME.button} px-4 py-2 rounded-full transition-all disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {isPending ? "Adding..." : "Add source"}
          </button>
        </>
      }
    >
      <fieldset>
        <legend className={THEME.form.label}>Type</legend>
        <div className="grid grid-cols-2 gap-2">
          {TYPE_CARDS.map((card) => {
            const selected = card.type === type;
            return (
              <label
                key={card.type}
                className={`cursor-pointer rounded-lg border p-3 transition-all ${selected ? "border-blue-300 bg-blue-50 ring-1 ring-blue-200" : "border-gray-200 hover:bg-gray-50"
                  }`}
              >
                <input
                  type="radio"
                  name="source-type"
                  value={card.type}
                  checked={selected}
                  onChange={() => chooseType(card.type)}
                  className="sr-only"
                />
                <span className={`flex items-center gap-1.5 text-xs font-black ${selected ? "text-blue-700" : "text-gray-700"}`}>
                  <span
                    aria-hidden="true"
                    className={`w-3 h-3 rounded-full border-2 ${selected ? "border-blue-600 bg-blue-600" : "border-gray-300"}`}
                  />
                  {card.title}
                  {card.type === "auto" && <LockClosedIcon className="w-3 h-3 text-gray-400" />}
                </span>
                <span className="block text-[11px] text-gray-500 mt-1 leading-snug">{card.description}</span>
              </label>
            );
          })}
        </div>
      </fieldset>

      <div>
        <label htmlFor={nameId} className={THEME.form.label}>
          {isAuto ? "System name" : "Source name"}
        </label>
        <input
          id={nameId}
          value={name}
          onChange={(e) => changeName(e.target.value)}
          placeholder={isAuto ? "e.g. JobSiteLK" : "e.g. Daily News"}
          autoComplete="off"
          spellCheck={!isAuto}
          className={`${THEME.form.input} ${isAuto ? "font-mono" : ""}`}
        />
        {taken && (
          <p className="text-[11px] font-bold text-red-500 mt-1.5">A source named &quot;{taken.name}&quot; already exists.</p>
        )}
        {!taken && caseOnlyMatch && (
          <p className="text-[11px] font-bold text-amber-700 mt-1.5">
            Similar to existing source &quot;{caseOnlyMatch.name}&quot;. Names are case-sensitive, so this would be a
            separate source.
          </p>
        )}
      </div>

      {isAuto && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 p-3 space-y-2 text-amber-900">
          <p className="font-bold">Must match the crawler exactly</p>
          <p className="text-[11px]">
            Use the exact name configured in the crawler (<code className="font-mono">SOURCE_NAME</code>),
            or the jobs it collects will be rejected.
          </p>
          <p>
            Will be saved as:{" "}
            <code className="font-mono font-bold text-gray-900 bg-white px-1.5 py-0.5 rounded border border-amber-200">
              {saved ? visibleWhitespace(name) : "—"}
            </code>
          </p>
          <ul className="list-disc pl-4 space-y-0.5 text-[11px]">
            <li>Case-sensitive: &quot;jobsitelk&quot; and &quot;JobSiteLK&quot; are different names.</li>
            {hasOuterWhitespace(name) && <li className="font-bold">Leading/trailing spaces will be removed.</li>}
            {hasInnerWhitespace(name) && (
              <li className="font-bold">Contains spaces (shown as ·). Check the crawler&apos;s name has them too.</li>
            )}
          </ul>
          <label htmlFor={confirmId} className="flex items-start gap-2 pt-1 cursor-pointer">
            <input
              id={confirmId}
              type="checkbox"
              checked={confirmed}
              onChange={(e) => setConfirmed(e.target.checked)}
              disabled={!saved}
              className="mt-0.5 cursor-pointer disabled:cursor-not-allowed"
            />
            <span className="font-semibold">I&apos;ve confirmed the crawler uses exactly this name.</span>
          </label>
        </div>
      )}

      <label className="flex items-center gap-2 cursor-pointer">
        <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} className="cursor-pointer" />
        <span className="font-semibold text-gray-700">{isAuto ? "Active from the next collection run" : "Active"}</span>
      </label>

      <p className="text-[11px] text-gray-400">Adding the name of a previously deleted source restores it.</p>

      {error && <p className="text-[11px] font-bold text-red-500">{error}</p>}
    </SourceDialog>
  );
}
