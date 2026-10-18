import { describe, expect, it } from 'vitest';

import { describeRetention, isEmpty, normalize, suggested } from './retention';

describe('retention', () => {
  it('knows an empty policy', () => {
    expect(isEmpty({})).toBe(true);
    expect(isEmpty({ keep_daily: 0, keep_within: '  ', prune: true })).toBe(true);
    expect(isEmpty({ keep_within: '30d' })).toBe(false);
    expect(isEmpty(suggested())).toBe(false);
  });

  it('normalizes form input', () => {
    // number inputs give strings or null when cleared
    const form = { keep_daily: '7', keep_weekly: null, keep_monthly: -2, keep_within: ' 1y ' } as unknown as Parameters<typeof normalize>[0];
    expect(normalize(form)).toEqual({
      keep_last: 0,
      keep_hourly: 0,
      keep_daily: 7,
      keep_weekly: 0,
      keep_monthly: 0,
      keep_yearly: 0,
      keep_within: '1y',
      prune: false,
    });
  });

  it('describes a policy', () => {
    expect(describeRetention(suggested())).toBe('Keep 7 daily, 4 weekly, 12 monthly, then prune');
    expect(describeRetention({ keep_last: 3, keep_within: '30d' })).toBe('Keep 3 last, everything from the last 30d');
    expect(describeRetention(null)).toBe('Keep everything');
  });
});
