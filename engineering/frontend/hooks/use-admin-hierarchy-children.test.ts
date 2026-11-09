import { describe, expect, it } from "vitest";
import { adminChildrenRequest } from "./use-admin-hierarchy-children";

describe("adminChildrenRequest", () => {
  it("builds the by-parent URL for an occupation level", () => {
    const req = adminChildrenRequest("occupation", 0, 4);
    expect(req.url).toBe("/api/v1/occupation/major-groups/4/sub-major-groups");
    expect(req.listField).toBe("sub_major_groups");
    expect(req.queryKey).toEqual(["occupation", "tree-children", "sub-major-groups", 4]);
  });

  it("builds the by-parent URL for the last industry level", () => {
    const req = adminChildrenRequest("industry", 3, 9);
    expect(req.url).toBe("/api/v1/industry/industry-classes/9/industry-subclasses");
    expect(req.listField).toBe("industry_subclasses");
  });
});
