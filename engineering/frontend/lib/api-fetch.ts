import { mockFetch } from "@/lib/mocks/mock-fetch";

/**
 * Whether requests made through `apiFetch` are answered by the in-memory mocks (lib/mocks)
 * instead of our /api routes. On unless `NEXT_PUBLIC_USE_MOCK_API=false`.
 *
 * To go live once the backend endpoints are deployed: set `NEXT_PUBLIC_USE_MOCK_API=false`
 * (or flip this default), then delete lib/mocks and this switch, and use `fetch` directly.
 * The hooks and components need no change.
 */
export const USE_MOCK_API = process.env.NEXT_PUBLIC_USE_MOCK_API !== "false";

/** `fetch`, or the mock transport while the backend isn't ready. Same signature and behaviour. */
export const apiFetch: typeof fetch = USE_MOCK_API ? mockFetch : (input, init) => fetch(input, init);
