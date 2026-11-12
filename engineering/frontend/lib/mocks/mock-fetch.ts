/**
 * MOCK TRANSPORT: answers the requests the hooks make to our /api/v1 routes from the in-memory
 * store in ./source-data, with the same URLs, methods, status codes and body shapes as the real
 * routes. Delete this folder once `USE_MOCK_API` is off (see lib/api-fetch.ts).
 */
import { ENV } from "@/config/env";
import type { CreateSourcePayload, SourceFilters, SourceType } from "@/types/source";
import {
  MockApiError,
  createSource,
  deleteSource,
  getSource,
  listSources,
  renameSource,
  updateSourceStatus,
} from "@/lib/mocks/source-data";

/** How long a mocked request takes, so loading and pending states are visible. */
const MOCK_LATENCY_MS = 350;

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

/** The error shape our /api routes forward from the backend: `getApiErrorMessage` reads `details.error`. */
function backendError(status: number, message: string): Response {
  return json({ error: "Backend returned an error", details: { error: message } }, status);
}

function wait(ms: number, signal?: AbortSignal | null): Promise<void> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) return reject(new DOMException("Aborted", "AbortError"));
    const timer = setTimeout(resolve, ms);
    signal?.addEventListener("abort", () => {
      clearTimeout(timer);
      reject(new DOMException("Aborted", "AbortError"));
    });
  });
}

function readBody<T>(init?: RequestInit): T {
  return (typeof init?.body === "string" ? JSON.parse(init.body) : {}) as T;
}

function filtersFrom(params: URLSearchParams): SourceFilters {
  const filters: SourceFilters = {};
  const type = params.get("type");
  if (type) filters.type = type as SourceType;
  const isActive = params.get("is-active");
  if (isActive !== null) filters.isActive = isActive === "true";
  return filters;
}

/** Routes one request to the store. Returns null when no mock handles it. */
function route(method: string, path: string, params: URLSearchParams, init?: RequestInit): Response | null {
  const base = `${ENV.NEXT_PUBLIC_API_BASE_URL}/sources`;

  if (path === base) {
    if (method === "GET") {
      const type = params.get("type");
      if (type && type !== "auto" && type !== "manual") {
        return backendError(400, `validation failed: invalid type "${type}", must be 'auto' or 'manual'`);
      }
      const sources = listSources(filtersFrom(params));
      return json({ count: sources.length, sources });
    }
    if (method === "POST") return json(createSource(readBody<CreateSourcePayload>(init)), 201);
  }

  const status = path.match(new RegExp(`^${base}/(\\d+)/status$`));
  if (status && method === "PATCH") {
    return json(updateSourceStatus(Number(status[1]), readBody<{ is_active: boolean }>(init).is_active));
  }

  const item = path.match(new RegExp(`^${base}/(\\d+)$`));
  if (item) {
    const id = Number(item[1]);
    if (method === "GET") return json(getSource(id));
    if (method === "PUT") return json(renameSource(id, readBody<{ name: string }>(init).name));
    if (method === "DELETE") {
      deleteSource(id);
      return json({ message: "Source deleted successfully" });
    }
  }

  return null;
}

/** A drop-in for `fetch` that serves the mocked endpoints. */
export async function mockFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  await wait(MOCK_LATENCY_MS, init?.signal);

  const url = new URL(typeof input === "string" ? input : input instanceof URL ? input.href : input.url, "http://mock.local");
  const method = (init?.method ?? "GET").toUpperCase();

  try {
    const response = route(method, url.pathname, url.searchParams, init);
    if (response) return response;
  } catch (err) {
    if (err instanceof MockApiError) return backendError(err.status, err.message);
    throw err;
  }

  console.warn(`[mock-fetch] no mock for ${method} ${url.pathname}`);
  return backendError(404, `No mock for ${method} ${url.pathname}`);
}
