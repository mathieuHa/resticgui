import { describe, expect, it } from 'vitest';

import { bytes, duration, elapsed, percent, relative } from './format';

describe('bytes', () => {
  it.each([
    [0, '0 B'],
    [1023, '1023 B'],
    [1536, '1.5 KiB'],
    [10 * 1024 * 1024, '10 MiB'],
    [123.4 * 1024 ** 3, '123 GiB'],
    [5 * 1024 ** 6, '5120 PiB'],
  ])('%d is %s', (n, want) => expect(bytes(n)).toBe(want));

  it('shows a dash for missing values', () => {
    expect(bytes(undefined)).toBe('–');
    expect(bytes(NaN)).toBe('–');
  });
});

describe('duration', () => {
  it.each([
    [0, '0s'],
    [42.4, '42s'],
    [60, '1m'],
    [125, '2m 5s'],
    [3725, '1h 2m'],
    [7200, '2h'],
    [90000, '1d 1h'],
  ])('%d seconds is %s', (s, want) => expect(duration(s)).toBe(want));

  it('shows a dash for missing or negative values', () => {
    expect(duration(null)).toBe('–');
    expect(duration(-1)).toBe('–');
  });
});

describe('elapsed', () => {
  it('measures between two times', () => {
    expect(elapsed('2026-10-18T10:00:00Z', '2026-10-18T10:01:30Z')).toBe(90);
  });
  it('measures until now while running', () => {
    expect(elapsed('2026-10-18T10:00:00Z', null, Date.parse('2026-10-18T10:00:10Z'))).toBe(10);
  });
  it('is null before the start', () => {
    expect(elapsed(null)).toBeNull();
  });
});

describe('relative', () => {
  const now = Date.parse('2026-10-18T12:00:00Z');
  it.each([
    ['2026-10-18T11:59:30Z', '30 seconds ago'],
    ['2026-10-18T11:55:00Z', '5 minutes ago'],
    ['2026-10-18T14:00:00Z', 'in 2 hours'],
    ['2026-10-17T12:00:00Z', 'yesterday'],
    ['2025-10-18T12:00:00Z', 'last year'],
  ])('%s is %s', (iso, want) => expect(relative(iso, now)).toBe(want));

  it('says never without a time', () => expect(relative(null)).toBe('never'));
});

describe('percent', () => {
  it('floors and clamps', () => {
    expect(percent(0.999)).toBe('99%');
    expect(percent(1.2)).toBe('100%');
    expect(percent(-1)).toBe('0%');
  });
});
