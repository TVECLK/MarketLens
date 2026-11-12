/**
 * Where a source's jobs come from:
 * - `auto`: a crawler collects them. The source's name must exactly match the crawler's
 *   SOURCE_NAME, or the jobs it sends are rejected.
 * - `manual`: entered by hand, e.g. newspaper ads.
 */
export const SOURCE_TYPES = ["auto", "manual"] as const;

export type SourceType = (typeof SOURCE_TYPES)[number];

/** How each type is named in the UI. */
export const SOURCE_TYPE_LABELS: Record<SourceType, string> = {
  auto: "Automatic",
  manual: "Manual",
};

export interface Source {
  id: number;
  name: string;
  type: SourceType;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/** GET /sources */
export interface SourceListResponse {
  count: number;
  sources: Source[];
}

/** Optional filters for GET /sources (`?type=` and `?is-active=`). */
export interface SourceFilters {
  type?: SourceType;
  isActive?: boolean;
}

/** POST /sources. `is_active` defaults to true on the backend. Re-using a deleted name restores it. */
export interface CreateSourcePayload {
  name: string;
  type: SourceType;
  is_active?: boolean;
}

/** PUT /sources/:id. Only manual sources can be renamed. */
export interface RenameSourcePayload {
  name: string;
}

/** PATCH /sources/:id/status */
export interface UpdateSourceStatusPayload {
  is_active: boolean;
}
