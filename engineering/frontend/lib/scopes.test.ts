import { describe, expect, it } from "vitest";
import { appScopes, CRUD_ACTIONS, PAGE_SCOPES, RESOURCE_SCOPES, scopeFor, scopeString } from "@/lib/scopes";

describe("scopeFor", () => {
  it("builds the permission string the backend checks", () => {
    expect(scopeFor(RESOURCE_SCOPES.experiences, CRUD_ACTIONS.update)).toBe("experiences:update");
  });
});

describe("appScopes", () => {
  it("requests the OIDC scopes, every action for every resource, and the page scopes", () => {
    const resources = Object.values(RESOURCE_SCOPES);
    const actions = Object.values(CRUD_ACTIONS);
    const pages = Object.values(PAGE_SCOPES);
    expect(appScopes.slice(0, 2)).toEqual(["openid", "profile"]);
    expect(appScopes).toHaveLength(2 + resources.length * actions.length + pages.length);
    for (const resource of resources) {
      expect(appScopes).toEqual(expect.arrayContaining(actions.map((action) => scopeFor(resource, action))));
    }
    expect(appScopes).toEqual(expect.arrayContaining(["kpis:access"]));
  });

  it("has no duplicates", () => {
    expect(new Set(appScopes).size).toBe(appScopes.length);
  });

  it("joins the scopes with spaces for the sign-in request", () => {
    expect(scopeString).toBe(appScopes.join(" "));
  });
});
