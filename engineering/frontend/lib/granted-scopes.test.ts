import { describe, expect, it } from "vitest";
import { hasScope, parseScopes, scopesFromAccessToken } from "@/lib/granted-scopes";
import { PAGE_SCOPES } from "@/lib/scopes";

/** An unsigned JWT with the given claims; only the payload matters here. */
const tokenWith = (claims: Record<string, unknown>) => {
  const base64url = (value: object) =>
    btoa(JSON.stringify(value)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  return `${base64url({ alg: "none" })}.${base64url(claims)}.`;
};

describe("parseScopes", () => {
  it("splits a space-separated string", () => {
    expect(parseScopes("openid  profile kpis:access")).toEqual(["openid", "profile", "kpis:access"]);
  });

  it("accepts a list", () => {
    expect(parseScopes(["openid", "kpis:access"])).toEqual(["openid", "kpis:access"]);
  });

  it("returns nothing for missing or odd values", () => {
    expect(parseScopes(undefined)).toEqual([]);
    expect(parseScopes(42)).toEqual([]);
  });
});

describe("scopesFromAccessToken", () => {
  it("reads the scope claim", () => {
    expect(scopesFromAccessToken(tokenWith({ scope: "openid kpis:access" }))).toEqual(["openid", "kpis:access"]);
  });

  it("reads a scopes list claim", () => {
    expect(scopesFromAccessToken(tokenWith({ scopes: ["openid"] }))).toEqual(["openid"]);
  });

  it("returns nothing without a readable token", () => {
    expect(scopesFromAccessToken(undefined)).toEqual([]);
    expect(scopesFromAccessToken("not-a-jwt")).toEqual([]);
    expect(scopesFromAccessToken("a.!!!.c")).toEqual([]);
  });
});

describe("hasScope", () => {
  it("matches the exact scope", () => {
    expect(hasScope(["openid", PAGE_SCOPES.kpis], PAGE_SCOPES.kpis)).toBe(true);
  });

  it("is false when the scope is missing", () => {
    expect(hasScope(["openid", "genders:create"], PAGE_SCOPES.kpis)).toBe(false);
  });

  it("lets a parent scope cover its children, like the SDK", () => {
    expect(hasScope(["kpis"], PAGE_SCOPES.kpis)).toBe(true);
    expect(hasScope(["kpi"], PAGE_SCOPES.kpis)).toBe(false);
  });
});
