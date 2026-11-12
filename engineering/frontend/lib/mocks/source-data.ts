/**
 * MOCK DATA: the only place dummy source data lives. Delete this folder once the backend's
 * /sources endpoints are deployed and `USE_MOCK_API` is off (see lib/api-fetch.ts).
 *
 * An in-memory store that follows the backend's rules, so the UI behaves as it will against the
 * real API: names are trimmed and unique (deleted ones included), creating a deleted name restores
 * it, only manual sources can be renamed, and deletes are soft. It lives for the browser session
 * (a page reload resets it).
 */
import type { CreateSourcePayload, Source, SourceFilters, SourceType } from "@/types/source";

interface StoredSource extends Source {
  deleted_at: string | null;
}

const SEED_TIME = "2026-11-09T00:00:00Z";

function seed(id: number, name: string, type: SourceType, isActive = true, deleted = false): StoredSource {
  return {
    id,
    name,
    type,
    is_active: isActive,
    created_at: SEED_TIME,
    updated_at: SEED_TIME,
    deleted_at: deleted ? SEED_TIME : null,
  };
}

/** Matches the backend seed (the five crawlers), plus a few manual sources to show every state. */
function initialSources(): StoredSource[] {
  return [
    seed(1, "Ikman", "auto"),
    seed(2, "XpressJobs", "auto"),
    seed(3, "TopJobs", "auto"),
    seed(4, "Rooster", "auto", false),
    seed(5, "GovernmentJobs", "auto"),
    seed(6, "Daily News", "manual"),
    seed(7, "Lankadeepa", "manual", false),
    // Deleted: adding a source named "Sunday Observer" restores this row.
    seed(8, "Sunday Observer", "manual", true, true),
  ];
}

let sources: StoredSource[] = initialSources();
let nextId = 9;

/** The error a request fails with, carrying the HTTP status the backend would return. */
export class MockApiError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message);
  }
}

function toSource(source: StoredSource): Source {
  const { id, name, type, is_active, created_at, updated_at } = source;
  return { id, name, type, is_active, created_at, updated_at };
}

function live(): StoredSource[] {
  return sources.filter((s) => s.deleted_at === null);
}

function findLive(id: number): StoredSource {
  const source = live().find((s) => s.id === id);
  if (!source) throw new MockApiError(404, "Failed to fetch source: not found");
  return source;
}

function touch(source: StoredSource) {
  source.updated_at = new Date().toISOString();
}

function assertNameFree(name: string, exceptId: number | null, action: string) {
  // The UNIQUE(name) constraint covers deleted rows too.
  if (sources.some((s) => s.name === name && s.id !== exceptId)) {
    throw new MockApiError(400, `Failed to ${action} source: already exists`);
  }
}

export function listSources(filters: SourceFilters = {}): Source[] {
  return live()
    .filter((s) => filters.type === undefined || s.type === filters.type)
    .filter((s) => filters.isActive === undefined || s.is_active === filters.isActive)
    .sort((a, b) => a.id - b.id)
    .map(toSource);
}

export function getSource(id: number): Source {
  return toSource(findLive(id));
}

export function createSource(payload: CreateSourcePayload): Source {
  const name = (payload.name ?? "").trim();
  if (!name) throw new MockApiError(400, "validation failed: name is required");
  if (payload.type !== "auto" && payload.type !== "manual") {
    throw new MockApiError(400, `validation failed: invalid type "${payload.type}", must be 'auto' or 'manual'`);
  }
  const isActive = payload.is_active ?? true;

  const deleted = sources.find((s) => s.name === name && s.deleted_at !== null);
  if (deleted) {
    deleted.type = payload.type;
    deleted.is_active = isActive;
    deleted.deleted_at = null;
    touch(deleted);
    return toSource(deleted);
  }

  assertNameFree(name, null, "create");
  const now = new Date().toISOString();
  const created: StoredSource = {
    id: nextId++,
    name,
    type: payload.type,
    is_active: isActive,
    created_at: now,
    updated_at: now,
    deleted_at: null,
  };
  sources.push(created);
  return toSource(created);
}

export function renameSource(id: number, rawName: string): Source {
  const name = (rawName ?? "").trim();
  if (!name) throw new MockApiError(400, "validation failed: name is required");
  const source = findLive(id);
  if (source.type !== "manual") {
    throw new MockApiError(
      400,
      `validation failed: only manual sources can be renamed; "${source.name}" is ${source.type} and must match its crawler's SOURCE_NAME`,
    );
  }
  if (source.name === name) return toSource(source);
  assertNameFree(name, id, "rename");
  source.name = name;
  touch(source);
  return toSource(source);
}

export function updateSourceStatus(id: number, isActive: boolean): Source {
  const source = findLive(id);
  source.is_active = isActive;
  touch(source);
  return toSource(source);
}

export function deleteSource(id: number): void {
  const source = findLive(id);
  source.deleted_at = new Date().toISOString();
  touch(source);
}

/** Restores the initial data. For tests. */
export function resetMockSources() {
  sources = initialSources();
  nextId = 9;
}
