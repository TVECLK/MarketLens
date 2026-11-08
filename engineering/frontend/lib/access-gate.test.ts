import { describe, expect, it } from "vitest";
import { ACCESS_DENIED_PATH, AUTH_PATH, gateDecision } from "@/lib/access-gate";
import { PAGE_SCOPES } from "@/lib/scopes";

const request = (overrides: Partial<Parameters<typeof gateDecision>[0]> = {}) => ({
  isSignedIn: true,
  grantedScopes: ["openid", "profile", PAGE_SCOPES.kpis],
  pathname: "/",
  isProtectedPath: false,
  hasAuthCode: false,
  ...overrides,
});

const signedOut = { isSignedIn: false, grantedScopes: [] };
const noAccess = { grantedScopes: ["openid", "profile", "genders:create"] };

describe("gateDecision", () => {
  it("lets signed-out visitors use public pages", () => {
    expect(gateDecision(request(signedOut))).toBe("allow");
  });

  it("sends signed-out visitors of protected pages to sign in", () => {
    expect(gateDecision(request({ ...signedOut, pathname: "/kpis", isProtectedPath: true }))).toBe("sign-in");
  });

  it("lets signed-in users with app access through, anywhere", () => {
    expect(gateDecision(request())).toBe("allow");
    expect(gateDecision(request({ pathname: "/kpis/gender", isProtectedPath: true }))).toBe("allow");
  });

  it("denies a signed-in user without app access, on every page", () => {
    expect(gateDecision(request(noAccess))).toBe("deny");
    expect(gateDecision(request({ ...noAccess, pathname: "/kpis", isProtectedPath: true }))).toBe("deny");
  });

  it("keeps the sign-out page reachable for a denied user", () => {
    expect(gateDecision(request({ ...noAccess, pathname: ACCESS_DENIED_PATH }))).toBe("allow");
  });

  describe("on the post-login page", () => {
    it("shows the loading page while the login completes", () => {
      expect(gateDecision(request({ ...signedOut, pathname: AUTH_PATH, hasAuthCode: true }))).toBe("allow");
    });

    it("goes home when opened with no login in progress", () => {
      expect(gateDecision(request({ ...signedOut, pathname: AUTH_PATH }))).toBe("to-home");
    });

    it("sends a user with access to the admin area", () => {
      expect(gateDecision(request({ pathname: AUTH_PATH }))).toBe("to-admin");
    });

    it("signs out a user without access", () => {
      expect(gateDecision(request({ ...noAccess, pathname: AUTH_PATH }))).toBe("deny");
    });
  });
});
