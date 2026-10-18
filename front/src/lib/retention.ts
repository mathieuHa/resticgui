import type { Retention } from './api/client';

export type CountRule = 'keep_last' | 'keep_hourly' | 'keep_daily' | 'keep_weekly' | 'keep_monthly' | 'keep_yearly';

export const COUNT_RULES: [CountRule, string][] = [
  ['keep_last', 'Last'],
  ['keep_hourly', 'Hourly'],
  ['keep_daily', 'Daily'],
  ['keep_weekly', 'Weekly'],
  ['keep_monthly', 'Monthly'],
  ['keep_yearly', 'Yearly'],
];

/** A policy that keeps a week of dailies, a month of weeklies and a year of monthlies. */
export const suggested = (): Retention => ({
  keep_last: 0,
  keep_hourly: 0,
  keep_daily: 7,
  keep_weekly: 4,
  keep_monthly: 12,
  keep_yearly: 0,
  keep_within: '',
  prune: true,
});

/** Without any rule restic would remove everything, so the server refuses it. */
export function isEmpty(r: Retention): boolean {
  return COUNT_RULES.every(([key]) => !r[key]) && !r.keep_within?.trim();
}

/** The policy as the API takes it: counts as numbers, blank fields as zero. */
export function normalize(r: Retention): Retention {
  const out: Retention = { keep_within: r.keep_within?.trim() ?? '', prune: !!r.prune };
  for (const [key] of COUNT_RULES) out[key] = Math.max(0, Math.floor(Number(r[key]) || 0));
  return out;
}

/** "7 daily, 4 weekly, everything from the last 30d" */
export function describeRetention(r: Retention | null | undefined): string {
  if (!r || isEmpty(r)) return 'Keep everything';
  const parts = COUNT_RULES.filter(([key]) => r[key]).map(([key, label]) => `${r[key]} ${label.toLowerCase()}`);
  if (r.keep_within) parts.push(`everything from the last ${r.keep_within}`);
  return `Keep ${parts.join(', ')}${r.prune ? ', then prune' : ''}`;
}
