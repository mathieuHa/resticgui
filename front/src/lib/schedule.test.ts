import { describe as group, expect, it } from 'vitest';

import { checkCron, defaults, describe, parse, toCron, type Preset } from './schedule';

group('schedule presets', () => {
  const cases: [string, Preset, string][] = [
    ['15 * * * *', { kind: 'hourly', minute: 15 }, 'Every hour at minute 15'],
    ['@every 6h', { kind: 'every', hours: 6 }, 'Every 6 hours'],
    ['30 2 * * *', { kind: 'daily', hour: 2, minute: 30 }, 'Every day at 02:30'],
    ['0 23 * * 5', { kind: 'weekly', weekday: 5, hour: 23, minute: 0 }, 'Every Friday at 23:00'],
    ['5 4 1 * *', { kind: 'monthly', day: 1, hour: 4, minute: 5 }, 'On the 1st of every month at 04:05'],
  ];

  it.each(cases)('%s round-trips', (cron, preset, sentence) => {
    expect(parse(cron)).toEqual(preset);
    expect(toCron(preset)).toBe(cron);
    expect(describe(preset)).toBe(sentence);
  });

  it.each([
    ['@daily', { kind: 'daily', hour: 0, minute: 0 }],
    ['@hourly', { kind: 'hourly', minute: 0 }],
    ['@weekly', { kind: 'weekly', weekday: 0, hour: 0, minute: 0 }],
    ['@monthly', { kind: 'monthly', day: 1, hour: 0, minute: 0 }],
    ['  30   2 * * * ', { kind: 'daily', hour: 2, minute: 30 }],
  ])('reads %s', (cron, preset) => expect(parse(cron)).toEqual(preset));

  it.each(['*/15 * * * *', '0 2 * * 1-5', '0 2 31 * *', '0 2 * 6 *', '@every 90m', '0 24 * * *', '@yearly'])(
    'keeps %s as custom',
    (cron) => {
      const p = parse(cron);
      expect(p).toEqual({ kind: 'custom', cron });
      expect(toCron(p)).toBe(cron);
    },
  );

  it('keeps the time when switching kind', () => {
    const daily: Preset = { kind: 'daily', hour: 3, minute: 45 };
    expect(defaults('weekly', daily)).toEqual({ kind: 'weekly', weekday: 0, hour: 3, minute: 45 });
    expect(defaults('custom', daily)).toEqual({ kind: 'custom', cron: '45 3 * * *' });
    expect(defaults('daily')).toEqual({ kind: 'daily', hour: 2, minute: 30 });
  });

  it('describes ordinals', () => {
    expect(describe({ kind: 'monthly', day: 2, hour: 0, minute: 0 })).toContain('2nd');
    expect(describe({ kind: 'monthly', day: 3, hour: 0, minute: 0 })).toContain('3rd');
    expect(describe({ kind: 'monthly', day: 11, hour: 0, minute: 0 })).toContain('11th');
    expect(describe({ kind: 'monthly', day: 21, hour: 0, minute: 0 })).toContain('21st');
  });
});

group('checkCron', () => {
  it.each(['30 2 * * *', '*/15 * * * 1-5', '@daily', '@every 6h', '@every 1h30m'])('accepts %s', (cron) =>
    expect(checkCron(cron)).toBeNull(),
  );
  it.each(['', '30 2 * *', '@sometimes', 'daily'])('refuses %s', (cron) => expect(checkCron(cron)).not.toBeNull());
});
