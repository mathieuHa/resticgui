// Schedules are stored as cron expressions; the UI edits them as presets.

export const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

export type Preset =
  | { kind: 'hourly'; minute: number }
  | { kind: 'every'; hours: number }
  | { kind: 'daily'; hour: number; minute: number }
  | { kind: 'weekly'; weekday: number; hour: number; minute: number }
  | { kind: 'monthly'; day: number; hour: number; minute: number }
  | { kind: 'custom'; cron: string };

export type PresetKind = Preset['kind'];

const int = (s: string | undefined, min: number, max: number): number | null => {
  if (s === undefined || !/^\d{1,2}$/.test(s)) return null;
  const n = Number(s);
  return n >= min && n <= max ? n : null;
};

/** Recognizes the cron expressions toCron produces; anything else is custom. */
export function parse(cron: string): Preset {
  const spec = cron.trim().replace(/\s+/g, ' ');
  const descriptors: Record<string, Preset> = {
    '@hourly': { kind: 'hourly', minute: 0 },
    '@daily': { kind: 'daily', hour: 0, minute: 0 },
    '@midnight': { kind: 'daily', hour: 0, minute: 0 },
    '@weekly': { kind: 'weekly', weekday: 0, hour: 0, minute: 0 },
    '@monthly': { kind: 'monthly', day: 1, hour: 0, minute: 0 },
  };
  const descriptor = descriptors[spec];
  if (descriptor) return descriptor;
  const every = /^@every (\d+)h$/.exec(spec);
  if (every) {
    const hours = Number(every[1]);
    if (hours >= 1 && hours <= 168) return { kind: 'every', hours };
  }

  const f = spec.split(' ');
  if (f.length === 5) {
    const [m, h, dom, mon, dow] = f;
    const minute = int(m, 0, 59);
    const hour = int(h, 0, 23);
    if (minute !== null && h === '*' && dom === '*' && mon === '*' && dow === '*') return { kind: 'hourly', minute };
    if (minute !== null && hour !== null && mon === '*') {
      if (dom === '*' && dow === '*') return { kind: 'daily', hour, minute };
      const weekday = int(dow, 0, 6);
      if (dom === '*' && weekday !== null) return { kind: 'weekly', weekday, hour, minute };
      const day = int(dom, 1, 28);
      if (day !== null && dow === '*') return { kind: 'monthly', day, hour, minute };
    }
  }
  return { kind: 'custom', cron: spec };
}

export function toCron(p: Preset): string {
  switch (p.kind) {
    case 'hourly':
      return `${p.minute} * * * *`;
    case 'every':
      return `@every ${p.hours}h`;
    case 'daily':
      return `${p.minute} ${p.hour} * * *`;
    case 'weekly':
      return `${p.minute} ${p.hour} * * ${p.weekday}`;
    case 'monthly':
      return `${p.minute} ${p.hour} ${p.day} * *`;
    case 'custom':
      return p.cron.trim();
  }
}

const pad = (n: number) => String(n).padStart(2, '0');
const time = (hour: number, minute: number) => `${pad(hour)}:${pad(minute)}`;
const ordinal = (n: number) => {
  const suffix = n % 10 === 1 && n !== 11 ? 'st' : n % 10 === 2 && n !== 12 ? 'nd' : n % 10 === 3 && n !== 13 ? 'rd' : 'th';
  return `${n}${suffix}`;
};

/** A sentence for the schedule, in the server's local time. */
export function describe(p: Preset): string {
  switch (p.kind) {
    case 'hourly':
      return p.minute === 0 ? 'Every hour' : `Every hour at minute ${p.minute}`;
    case 'every':
      return p.hours === 1 ? 'Every hour' : `Every ${p.hours} hours`;
    case 'daily':
      return `Every day at ${time(p.hour, p.minute)}`;
    case 'weekly':
      return `Every ${WEEKDAYS[p.weekday]} at ${time(p.hour, p.minute)}`;
    case 'monthly':
      return `On the ${ordinal(p.day)} of every month at ${time(p.hour, p.minute)}`;
    case 'custom':
      return `Cron: ${p.cron}`;
  }
}

/** The preset a kind starts with when the user switches to it. */
export function defaults(kind: PresetKind, from?: Preset): Preset {
  const hour = from && 'hour' in from ? from.hour : 2;
  const minute = from && 'minute' in from ? from.minute : 30;
  switch (kind) {
    case 'hourly':
      return { kind, minute };
    case 'every':
      return { kind, hours: 6 };
    case 'daily':
      return { kind, hour, minute };
    case 'weekly':
      return { kind, weekday: 0, hour, minute };
    case 'monthly':
      return { kind, day: 1, hour, minute };
    case 'custom':
      return { kind, cron: from ? toCron(from) : `${minute} ${hour} * * *` };
  }
}

/**
 * A quick check before the server validates it: five fields or a descriptor.
 * Returns an error message or null.
 */
export function checkCron(cron: string): string | null {
  const spec = cron.trim();
  if (spec === '') return 'Enter a schedule';
  if (spec.startsWith('@')) {
    return /^@(yearly|annually|monthly|weekly|daily|midnight|hourly|every \d+[smh](\d+[smh])*)$/.test(spec)
      ? null
      : 'Use a descriptor such as @daily or @every 6h';
  }
  return spec.split(/\s+/).length === 5 ? null : 'A cron expression has five fields: minute hour day month weekday';
}
