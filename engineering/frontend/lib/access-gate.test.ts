import { describe, expect, it } from "vitest";
import {
  ACCESS_DENIED_PATH,
  AUTH_PATH,
  adminHomePath,
  gateDecision,
  hasAppAccess,
  requiredPageScope,
} from "@/lib/access-gate";
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

  describe("with page scopes (any one lets a user in, each page needs its own)", () => {
    const kpisOnly = { grantedScopes: ["openid", "profile", PAGE_SCOPES.kpis] };
    const dataSourcesOnly = { grantedScopes: ["openid", "profile", PAGE_SCOPES.datasource] };
    const both = { grantedScopes: ["openid", "profile", PAGE_SCOPES.kpis, PAGE_SCOPES.datasource] };

    it("lets a Data Sources-only user into the app instead of denying them", () => {
      expect(gateDecision(request({ ...dataSourcesOnly, pathname: "/" }))).toBe("allow");
      expect(gateDecision(request({ ...dataSourcesOnly, pathname: "/data-sources", isProtectedPath: true }))).toBe("allow");
      expect(gateDecision(request({ ...dataSourcesOnly, pathname: AUTH_PATH }))).toBe("to-admin");
    });

    it("sends a user to their admin home from a page they don't have the scope for", () => {
      expect(gateDecision(request({ ...dataSourcesOnly, pathname: "/kpis", isProtectedPath: true }))).toBe("to-admin");
      expect(gateDecision(request({ ...dataSourcesOnly, pathname: "/kpis/gender", isProtectedPath: true }))).toBe("to-admin");
      expect(gateDecision(request({ ...kpisOnly, pathname: "/data-sources", isProtectedPath: true }))).toBe("to-admin");
      expect(gateDecision(request({ ...kpisOnly, pathname: "/data-sources/x", isProtectedPath: true }))).toBe("to-admin");
    });

    it("lets a user with both scopes open both pages", () => {
      expect(gateDecision(request({ ...both, pathname: "/kpis", isProtectedPath: true }))).toBe("allow");
      expect(gateDecision(request({ ...both, pathname: "/data-sources", isProtectedPath: true }))).toBe("allow");
    });

    it("doesn't treat a similar-looking path as an admin page", () => {
      expect(gateDecision(request({ pathname: "/data-sources-old" }))).toBe("allow");
    });
  });
});

describe("hasAppAccess", () => {
  it("is true with any page scope and false with none", () => {
    expect(hasAppAccess([PAGE_SCOPES.kpis])).toBe(true);
    expect(hasAppAccess([PAGE_SCOPES.datasource])).toBe(true);
    expect(hasAppAccess(["openid", "genders:create"])).toBe(false);
  });
});

describe("requiredPageScope", () => {
  it("names each admin page's scope, for the page and below, and none elsewhere", () => {
    expect(requiredPageScope("/kpis")).toBe(PAGE_SCOPES.kpis);
    expect(requiredPageScope("/kpis/gender")).toBe(PAGE_SCOPES.kpis);
    expect(requiredPageScope("/data-sources")).toBe(PAGE_SCOPES.datasource);
    expect(requiredPageScope("/data-sources/1")).toBe(PAGE_SCOPES.datasource);
    expect(requiredPageScope("/")).toBeUndefined();
    expect(requiredPageScope("/data-sources-old")).toBeUndefined();
  });
});

describe("adminHomePath", () => {
  it("is the first admin page the user can open", () => {
    expect(adminHomePath([PAGE_SCOPES.kpis, PAGE_SCOPES.datasource])).toBe("/kpis");
    expect(adminHomePath([PAGE_SCOPES.kpis])).toBe("/kpis");
    expect(adminHomePath([PAGE_SCOPES.datasource])).toBe("/data-sources");
    expect(adminHomePath(["openid"])).toBeUndefined();
  });

  it("is always a page the gate lets the user open, so redirecting there can't loop", () => {
    for (const scopes of [[PAGE_SCOPES.kpis], [PAGE_SCOPES.datasource], [PAGE_SCOPES.kpis, PAGE_SCOPES.datasource]]) {
      const home = adminHomePath(scopes)!;
      expect(gateDecision(request({ grantedScopes: scopes, pathname: home, isProtectedPath: true }))).toBe("allow");
    }
  });
});

describe("gateDecision on the post-login page", () => {
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
