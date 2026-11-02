/**
 * True when every item's value is zero (or there are no items). The backend
 * returns every category even for a range with no jobs, with counts of 0, so
 * "no data" means all zeros, not an empty list.
 */
export function isEmptyChartData<T>(data: readonly T[], getValue: (item: T) => number): boolean {
  return data.every((item) => getValue(item) === 0);
}
