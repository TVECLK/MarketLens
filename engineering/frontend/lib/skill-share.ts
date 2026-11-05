/**
 * A skill's share of the vacancies in the selected occupation, as a whole
 * percent. `skillVacancies` and `totalVacancies` are both sums of
 * `no_of_vacancies` over the occupation's job posts, so the skill can't
 * exceed the total; the cap is only a safety net. Null when the total is
 * unknown (its request failed) or zero.
 */
export function skillSharePercent(skillVacancies: number, totalVacancies: number | null): number | null {
  if (!totalVacancies || totalVacancies <= 0) return null;
  return Math.min(100, Math.round((skillVacancies / totalVacancies) * 100));
}
