const UNITS = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB'];

/** 1536 → "1.5 KiB" */
export function bytes(n: number | undefined | null): string {
  if (n === undefined || n === null || !Number.isFinite(n)) return '–';
  let value = n;
  let unit = 0;
  while (Math.abs(value) >= 1024 && unit < UNITS.length - 1) {
    value /= 1024;
    unit++;
  }
  const digits = unit === 0 || value >= 100 ? 0 : value >= 10 ? 1 : 2;
  return `${Number(value.toFixed(digits))} ${UNITS[unit]}`;
}

/** 3725 → "1h 2m", 42 → "42s" */
export function duration(seconds: number | undefined | null): string {
  if (seconds === undefined || seconds === null || !Number.isFinite(seconds) || seconds < 0) return '–';
  const s = Math.round(seconds);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  if (m < 60) return s % 60 ? `${m}m ${s % 60}s` : `${m}m`;
  const h = Math.floor(m / 60);
  if (h < 24) return m % 60 ? `${h}h ${m % 60}m` : `${h}h`;
  const d = Math.floor(h / 24);
  return h % 24 ? `${d}d ${h % 24}h` : `${d}d`;
}

/** Seconds between two ISO times, or until now when end is missing. */
export function elapsed(start: string | null | undefined, end?: string | null, now = Date.now()): number | null {
  if (!start) return null;
  const from = Date.parse(start);
  const to = end ? Date.parse(end) : now;
  return Number.isNaN(from) || Number.isNaN(to) ? null : Math.max(0, (to - from) / 1000);
}

const relativeFormat = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });
const STEPS: [Intl.RelativeTimeFormatUnit, number][] = [
  ['second', 60],
  ['minute', 60],
  ['hour', 24],
  ['day', 30],
  ['month', 12],
  ['year', Infinity],
];

/** "5 minutes ago", "in 2 hours" */
export function relative(iso: string | null | undefined, now = Date.now()): string {
  if (!iso) return 'never';
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return '–';
  let value = (t - now) / 1000;
  for (const [unit, size] of STEPS) {
    if (Math.abs(value) < size) return relativeFormat.format(Math.round(value), unit);
    value /= size;
  }
  return '–';
}

const dateTimeFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });

export function dateTime(iso: string | null | undefined): string {
  if (!iso) return '–';
  const t = Date.parse(iso);
  return Number.isNaN(t) ? '–' : dateTimeFormat.format(t);
}

export function percent(fraction: number | undefined): string {
  if (fraction === undefined || !Number.isFinite(fraction)) return '–';
  return `${Math.floor(Math.min(1, Math.max(0, fraction)) * 100)}%`;
}
