import type { Job, Paths, Plan, PlanInput, PlanOptions, Repository, Retention } from './api/client';
import { isEmpty, normalize, suggested } from './retention';

/** A plan as its form edits it. */
export type PlanDraft = {
  name: string;
  paths: Paths;
  /** more tags for its snapshots, separated by spaces or commas */
  tags: string;
  /** null when backups only run by hand */
  backupCron: string | null;
  retentionOn: boolean;
  retention: Retention;
  forgetCron: string | null;
  compression: PlanOptions['compression'];
  exclude_larger_than: string;
  one_file_system: boolean;
  exclude_caches: boolean;
  /** file names, one per line */
  exclude_if_present: string;
  /** KiB/s, null or 0 for no limit */
  limit_upload: number | null;
  limit_download: number | null;
};

export const DEFAULT_PLAN = 'Default';

export function newPlanDraft(name = DEFAULT_PLAN): PlanDraft {
  return {
    name,
    paths: { include: [], exclude: [] },
    tags: '',
    backupCron: '30 2 * * *',
    retentionOn: true,
    retention: suggested(),
    forgetCron: '0 4 * * 0',
    compression: 'auto',
    exclude_larger_than: '',
    one_file_system: false,
    exclude_caches: false,
    exclude_if_present: '',
    limit_upload: null,
    limit_download: null,
  };
}

export function draftOfPlan(p: Plan): PlanDraft {
  const cron = (job: 'backup' | 'forget') => {
    const s = p.schedules.find((x) => x.job === job);
    return s?.enabled ? s.cron : null;
  };
  const o = p.options;
  return {
    name: p.name,
    paths: { include: [...p.paths.include], exclude: [...p.paths.exclude] },
    tags: p.tags.join(' '),
    backupCron: cron('backup'),
    retentionOn: !!p.retention,
    retention: p.retention ? { ...p.retention } : suggested(),
    forgetCron: cron('forget'),
    compression: o.compression,
    exclude_larger_than: o.exclude_larger_than,
    one_file_system: o.one_file_system,
    exclude_caches: o.exclude_caches,
    exclude_if_present: o.exclude_if_present.join('\n'),
    limit_upload: o.limit_upload || null,
    limit_download: o.limit_download || null,
  };
}

const words = (text: string, sep: RegExp) =>
  text
    .split(sep)
    .map((w) => w.trim())
    .filter(Boolean);

/** The request body for a draft. */
export function toPlanInput(d: PlanDraft): PlanInput {
  const schedules: PlanInput['schedules'] = [];
  if (d.backupCron) schedules.push({ job: 'backup', cron: d.backupCron, enabled: true });
  if (d.retentionOn && d.forgetCron) schedules.push({ job: 'forget', cron: d.forgetCron, enabled: true });
  return {
    name: d.name.trim(),
    tags: words(d.tags, /[\s,]+/),
    paths: { include: [...d.paths.include], exclude: [...d.paths.exclude] },
    options: {
      compression: d.compression,
      exclude_larger_than: d.exclude_larger_than.trim(),
      one_file_system: d.one_file_system,
      exclude_caches: d.exclude_caches,
      exclude_if_present: words(d.exclude_if_present, /\n/),
      limit_upload: Math.max(0, Math.floor(Number(d.limit_upload) || 0)),
      limit_download: Math.max(0, Math.floor(Number(d.limit_download) || 0)),
    },
    schedules,
    retention: d.retentionOn ? normalize(d.retention) : null,
  };
}

/** Problems the server would refuse, found before sending. */
export function planProblems(d: PlanDraft, others: string[] = []): string[] {
  const out: string[] = [];
  const name = d.name.trim();
  if (!name) out.push('Give the plan a name');
  else if (others.includes(name)) out.push(`Another plan is already named ${name}`);
  if (!d.paths.include.length) out.push('Add at least one folder or file to back up');
  if (d.retentionOn && isEmpty(d.retention)) out.push('Set at least one retention rule, or turn retention off');
  if (words(d.tags, /[\s,]+/).some((t) => t.startsWith('resticgui:'))) out.push('Tags starting with resticgui: are ResticGUI’s own');
  if (d.exclude_larger_than.trim() && !/^[0-9]+[kKmMgGtT]?$/.test(d.exclude_larger_than.trim()))
    out.push('Skip files larger than takes a size such as 500M or 2G');
  return out;
}

/** The plan a job ran, when it still exists. */
export function planOf(repo: Repository, job: Pick<Job, 'plan_id'>): Plan | undefined {
  return job.plan_id == null ? undefined : repo.plans.find((p) => p.id === job.plan_id);
}

/** The plans that have something to back up. */
export function ready(repo: Repository): Plan[] {
  return repo.plans.filter((p) => p.paths.include.length > 0);
}

/** When the next scheduled backup of any plan runs, if one is scheduled. */
export function nextBackup(repo: Repository): string | undefined {
  return repo.plans
    .flatMap((p) => p.schedules)
    .filter((s) => s.job === 'backup' && s.next_run)
    .map((s) => s.next_run!)
    .sort((a, b) => Date.parse(a) - Date.parse(b))[0];
}

/** The plan whose snapshot this is, from its tags. */
export function planOfSnapshot(repo: Repository, tags: string[] | undefined): Plan | undefined {
  return repo.plans.find((p) => tags?.includes(p.tag));
}
