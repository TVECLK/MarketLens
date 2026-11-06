import { describe, expect, it } from "vitest";
import { skillSharePercent } from "./skill-share";

describe("skillSharePercent", () => {
  it("returns the skill's share of all vacancies as a rounded percent", () => {
    expect(skillSharePercent(25, 100)).toBe(25);
    expect(skillSharePercent(1, 3)).toBe(33);
    expect(skillSharePercent(2, 3)).toBe(67);
  });

  it("does not depend on how many skills are loaded", () => {
    expect(skillSharePercent(50, 200)).toBe(25);
  });

  it("caps at 100", () => {
    expect(skillSharePercent(120, 100)).toBe(100);
  });

  it("returns null when the total is unknown or zero", () => {
    expect(skillSharePercent(5, null)).toBeNull();
    expect(skillSharePercent(5, 0)).toBeNull();
  });
});
