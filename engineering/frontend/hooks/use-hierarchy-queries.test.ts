import { describe, expect, it } from "vitest";
import {
  hierarchyAnalysisRequest,
  hierarchyChildrenRequest,
  hierarchyRootRequest,
} from "./use-hierarchy-queries";

const FROM = "2026-01-01";
const TO = "2026-01-31";
const DATES = "from-date=2026-01-01&to-date=2026-01-31";

describe("hierarchyRootRequest", () => {
  it("builds the URL and query key for the top level", () => {
    const req = hierarchyRootRequest("occupation", "major-groups", FROM, TO);
    expect(req.url).toBe(`/api/v1/occupation/major-groups?${DATES}`);
    expect(req.queryKey).toEqual(["occupation", "major-groups", FROM, TO]);
  });

  it("is enabled only when both dates are set", () => {
    expect(hierarchyRootRequest("industry", "industry-sectors", FROM, TO).enabled).toBe(true);
    expect(hierarchyRootRequest("industry", "industry-sectors", "", TO).enabled).toBe(false);
    expect(hierarchyRootRequest("industry", "industry-sectors", FROM, "").enabled).toBe(false);
  });

  it("is disabled for a reversed date range", () => {
    expect(hierarchyRootRequest("industry", "industry-sectors", TO, FROM).enabled).toBe(false);
  });
});

describe("hierarchyChildrenRequest", () => {
  it("builds the URL and query key for the children of a parent", () => {
    const req = hierarchyChildrenRequest("industry", "industry-sectors", "industry-divisions", 7, FROM, TO);
    expect(req.url).toBe(`/api/v1/industry/industry-sectors/7/industry-divisions?${DATES}`);
    expect(req.queryKey).toEqual(["industry", "industry-divisions", 7, FROM, TO]);
  });

  it("is enabled only with a parent id and both dates", () => {
    const enabled = (parentId: number | null, from = FROM, to = TO) =>
      hierarchyChildrenRequest("occupation", "major-groups", "sub-major-groups", parentId, from, to).enabled;
    expect(enabled(3)).toBe(true);
    expect(enabled(null)).toBe(false);
    expect(enabled(3, "")).toBe(false);
    expect(enabled(3, FROM, "")).toBe(false);
    expect(enabled(3, TO, FROM)).toBe(false);
  });
});

describe("hierarchyAnalysisRequest", () => {
  it("puts the standard, level and id in the URL and query key", () => {
    const req = hierarchyAnalysisRequest("occupation", "minor-group", 12, FROM, TO);
    expect(req.url).toBe(`/api/v1/analysis/occupation/minor-group/12?${DATES}`);
    expect(req.queryKey).toEqual(["analysis", "occupation", "minor-group", 12, FROM, TO]);
  });

  it("uses a different query key for each standard", () => {
    const occupation = hierarchyAnalysisRequest("occupation", "unit-group", 1, FROM, TO);
    const industry = hierarchyAnalysisRequest("industry", "unit-group", 1, FROM, TO);
    expect(occupation.queryKey).not.toEqual(industry.queryKey);
  });

  it("is enabled only with a level, an id and a valid date range", () => {
    const enabled = (level: string | null, id: number | null, from = FROM, to = TO) =>
      hierarchyAnalysisRequest("industry", level, id, from, to).enabled;
    expect(enabled("industry-group", 4)).toBe(true);
    expect(enabled(null, 4)).toBe(false);
    expect(enabled("industry-group", null)).toBe(false);
    expect(enabled("industry-group", 4, "")).toBe(false);
    expect(enabled("industry-group", 4, TO, FROM)).toBe(false);
  });
});
