import type { Job, Plan, Repository } from './api/client';
import { isActive } from './jobs.svelte';
import { ready } from './plan';
import { parse, type Preset } from './schedule';

/** How a repository is doing, worst first in DASHBOARD_ORDER. */
export type Health = 'failed' | 'late' | 'setup' | 'never' | 'running' | 'ok';

export const DASHBOARD_ORDER: Health[] = ['failed', 'late', 'setup', 'never', 'running', 'ok'];

/** States that need the user to do something. */
export const NEEDS_ACTION: Health[] = ['failed', 'late', 'setup', 'never'];

const HOUR = 3600_000;

/** How often a preset runs, in hours; custom cron gets a week. */
export function periodHours(p: Preset): number {
  switch (p.kind) {
    case 'hourly':
      return 1;
    case 'every':
      return p.hours;
    case 'daily':
      return 24;
    case 'weekly':
      return 168;
    case 'monthly':
      return 31 * 24;
    case 'custom':
      return 168;
  }
}

/**
 * The state of a plan from the jobs of its repository (any order). A
 * scheduled backup is late when the last good one is older than one and a
 * half periods, plus an hour for a backup that is slow to finish.
 */
export function planHealth(plan: Plan, jobs: Job[], now = Date.now()): Health {
  const backups = jobs
    .filter((j) => j.kind === 'backup' && j.plan_id === plan.id && j.status !== 'cancelled' && j.finished_at)
    .sort((a, b) => Date.parse(b.finished_at!) - Date.parse(a.finished_at!));
  const last = backups[0];
  if (last && (last.status === 'failed' || last.status === 'interrupted')) return 'failed';
  const good = backups.find((j) => j.status === 'succeeded');
  if (!good) return 'never';
  const schedule = plan.schedules.find((s) => s.job === 'backup' && s.enabled);
  if (schedule) {
    const allowed = periodHours(parse(schedule.cron)) * 1.5 * HOUR + HOUR;
    if (now - Date.parse(good.finished_at!) > allowed) return 'late';
  }
  return 'ok';
}

/** The state of a repository: the worst of its plans'. */
export function health(repo: Repository, jobs: Job[], now = Date.now()): Health {
  if (jobs.some((j) => isActive(j.status))) return 'running';
  const plans = ready(repo);
  if (!repo.initialized_at || plans.length === 0) return 'setup';
  const states = plans.map((p) => planHealth(p, jobs, now));
  return DASHBOARD_ORDER.find((h) => states.includes(h)) ?? 'ok';
}

/** One sentence for the top of the dashboard. */
export function headline(states: Health[]): { tone: 'success' | 'warning' | 'error' | 'info'; text: string } {
  const count = (h: Health) => states.filter((s) => s === h).length;
  const plural = (n: number, one: string, many: string) => (n === 1 ? one : many);
  if (count('failed')) {
    const n = count('failed');
    return { tone: 'error', text: `${n} ${plural(n, 'repository has', 'repositories have')} a failed backup` };
  }
  const waiting = states.filter((s) => NEEDS_ACTION.includes(s)).length;
  if (waiting) return { tone: 'warning', text: `${waiting} ${plural(waiting, 'repository needs', 'repositories need')} your attention` };
  if (count('running')) return { tone: 'info', text: 'A backup is running; everything else is up to date' };
  const n = states.length;
  return { tone: 'success', text: n === 1 ? 'Your repository is backed up' : `All ${n} repositories are backed up` };
}
