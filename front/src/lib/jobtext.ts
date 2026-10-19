import type { Job, JobKind } from './api/client';
import { bytes, duration } from './format';

export const KIND_LABELS: Record<JobKind, string> = {
  init: 'Create repository',
  backup: 'Backup',
  forget: 'Apply retention',
  prune: 'Prune',
  check: 'Check',
  unlock: 'Unlock',
  restore: 'Restore',
  repair_index: 'Rebuild index',
  repair_snapshots: 'Repair snapshots',
  migrate: 'Upgrade format',
  change_password: 'Change password',
};

/** How much data a check reads, as restic's --read-data-subset, and how it is said. */
export const SUBSETS: [string, string][] = [
  ['', 'No data, only the structure'],
  ['5%', '5% of the data'],
  ['10%', '10% of the data'],
  ['25%', '25% of the data'],
  ['100%', 'All the data'],
];

/** "10% of the data read", or '' for a check of the structure only. */
export function subsetText(subset: unknown): string {
  if (typeof subset !== 'string' || !subset) return '';
  if (subset === '100%') return 'all the data read';
  return `${subset}${/^[0-9.]+%$/.test(subset) ? ' of the data' : ''} read`;
}

type Summary = Record<string, unknown>;
const num = (s: Summary, key: string) => (typeof s[key] === 'number' ? (s[key] as number) : undefined);

const plural = (n: number, word: string) => `${n} ${word}${n === 1 ? '' : 's'}`;

/** One line about what a finished job did, from restic's summary. */
export function summarize(job: Job): string {
  const s = (job.summary ?? {}) as Summary;
  switch (job.kind) {
    case 'backup': {
      const added = num(s, 'data_added');
      if (added === undefined) return '';
      const parts = [
        `${num(s, 'files_new') ?? 0} new`,
        `${num(s, 'files_changed') ?? 0} changed`,
        `${num(s, 'files_unmodified') ?? 0} unchanged files`,
        `${bytes(added)} added`,
      ];
      const took = num(s, 'total_duration');
      if (took !== undefined) parts.push(`in ${duration(took)}`);
      return parts.join(', ');
    }
    case 'forget': {
      const kept = num(s, 'kept');
      if (kept === undefined) return '';
      const text = `Kept ${plural(kept, 'snapshot')}, removed ${num(s, 'removed') ?? 0}`;
      return s.pruned ? `${text}, then pruned` : text;
    }
    case 'check': {
      const read = subsetText((job.params as Summary | undefined)?.read_data_subset);
      const errors = num(s, 'num_errors');
      let text = errors === undefined ? (job.status === 'succeeded' ? 'No errors found' : '') : errors === 0 ? 'No errors found' : plural(errors, 'error');
      if (errors && s.suggest_repair_index) text += ': rebuild the index, in Settings';
      return text && read ? `${text}, ${read}` : text;
    }
    case 'repair_snapshots': {
      const modified = num(s, 'modified');
      if (modified === undefined) return '';
      return modified === 0 ? 'No snapshot needed repair' : `Repaired ${plural(modified, 'snapshot')}`;
    }
    case 'migrate': {
      if (!Array.isArray(s.applied)) return '';
      return s.applied.length ? 'Upgraded to format version 2' : 'Already at the latest format';
    }
    case 'restore': {
      const files = num(s, 'files_restored');
      const size = num(s, 'bytes_restored');
      const target = (job.params as Summary | undefined)?.target;
      if (files === undefined) return typeof target === 'string' ? `Into ${target}` : '';
      return `${plural(files, 'file')} (${bytes(size)})${typeof target === 'string' ? ` into ${target}` : ''}`;
    }
    default:
      return '';
  }
}
