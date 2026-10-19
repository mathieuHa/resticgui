import { describe, expect, it } from 'vitest';

import type { Job } from './api/client';
import { summarize } from './jobtext';

const job = (kind: Job['kind'], summary: Record<string, unknown>, params: Record<string, unknown> = {}): Job => ({
  id: 1,
  repository_id: 1,
  plan_id: null,
  kind,
  trigger: 'manual',
  status: 'succeeded',
  created_at: '',
  started_at: null,
  finished_at: null,
  summary,
  params,
});

describe('summarize', () => {
  it('describes a backup', () => {
    expect(
      summarize(job('backup', { files_new: 3, files_changed: 1, files_unmodified: 10, data_added: 2048, total_duration: 75 })),
    ).toBe('3 new, 1 changed, 10 unchanged files, 2 KiB added, in 1m 15s');
  });
  it('describes a forget', () => {
    expect(summarize(job('forget', { kept: 1, removed: 4, pruned: true }))).toBe('Kept 1 snapshot, removed 4, then pruned');
    expect(summarize(job('forget', { kept: 5, removed: 0, pruned: false }))).toBe('Kept 5 snapshots, removed 0');
  });
  it('describes a check', () => {
    expect(summarize(job('check', { num_errors: 0 }))).toBe('No errors found');
    expect(summarize(job('check', { num_errors: 2 }))).toBe('2 errors');
  });
  it('describes a restore', () => {
    expect(summarize(job('restore', { files_restored: 1, bytes_restored: 10 }, { target: '/tmp/r' }))).toBe(
      '1 file (10 B) into /tmp/r',
    );
    expect(summarize(job('restore', {}, { target: '/tmp/r' }))).toBe('Into /tmp/r');
  });
  it('says nothing without a summary', () => {
    expect(summarize(job('backup', {}))).toBe('');
    expect(summarize(job('unlock', {}))).toBe('');
  });
});
