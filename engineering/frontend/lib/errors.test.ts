import { describe, expect, it } from "vitest";
import { getApiErrorMessage, getErrorMessage } from "./errors";

describe("getErrorMessage", () => {
  it("returns the message of an Error", () => {
    expect(getErrorMessage(new Error("boom"), "fallback")).toBe("boom");
  });

  it("returns the fallback for non-Error values", () => {
    expect(getErrorMessage("boom", "fallback")).toBe("fallback");
    expect(getErrorMessage(null)).toBeUndefined();
  });
});

describe("getApiErrorMessage", () => {
  const FALLBACK = "Failed to create gender";

  it("unwraps the backend's error from the route envelope", () => {
    const body = {
      error: "Backend returned an error",
      details: { error: "missing required scope: genders:create" },
    };
    expect(getApiErrorMessage(body, FALLBACK)).toBe("missing required scope: genders:create");
  });

  it("does not expose the backend's raw details", () => {
    const body = {
      error: "Backend returned an error",
      details: { error: "Failed to create gender", details: "pq: duplicate key value" },
    };
    expect(getApiErrorMessage(body, FALLBACK)).toBe("Failed to create gender");
  });

  it("parses backend bodies forwarded as JSON text", () => {
    const body = {
      error: "Backend returned an error",
      details: '{"error":"Failed to fetch genders"}',
    };
    expect(getApiErrorMessage(body, FALLBACK)).toBe("Failed to fetch genders");
  });

  it("uses the route's own error when there is no backend body", () => {
    expect(getApiErrorMessage({ error: "Please sign in first" }, FALLBACK)).toBe("Please sign in first");
    expect(getApiErrorMessage({ error: "Failed to create gender", details: "fetch failed" }, FALLBACK)).toBe(
      "Failed to create gender",
    );
  });

  it("falls back for empty or non-object bodies", () => {
    expect(getApiErrorMessage(null, FALLBACK)).toBe(FALLBACK);
    expect(getApiErrorMessage("<html>", FALLBACK)).toBe(FALLBACK);
    expect(getApiErrorMessage({}, FALLBACK)).toBe(FALLBACK);
  });
});
