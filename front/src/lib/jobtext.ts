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
};

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
      const errors = num(s, 'num_errors');
      if (errors === undefined) return job.status === 'succeeded' ? 'No errors found' : '';
      return errors === 0 ? 'No errors found' : plural(errors, 'error');
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
