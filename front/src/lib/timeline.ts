import type { Job } from './api/client';

export type Day = { key: string; label: string; jobs: Job[] };

const dayFormat = new Intl.DateTimeFormat(undefined, { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });

function dayKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

/** Jobs grouped by the local day they started (or were queued), newest first. */
export function byDay(jobs: Job[], now = new Date()): Day[] {
  const today = dayKey(now);
  const yesterday = dayKey(new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1));
  const when = (j: Job) => Date.parse(j.started_at ?? j.created_at);
  const days = new Map<string, Day>();
  for (const job of [...jobs].sort((a, b) => when(b) - when(a) || b.id - a.id)) {
    const d = new Date(when(job));
    const key = dayKey(d);
    let day = days.get(key);
    if (!day) {
      day = { key, label: key === today ? 'Today' : key === yesterday ? 'Yesterday' : dayFormat.format(d), jobs: [] };
      days.set(key, day);
    }
    day.jobs.push(job);
  }
  return [...days.values()];
}
