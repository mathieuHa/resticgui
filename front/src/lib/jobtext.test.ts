import { describe, expect, it } from 'vitest';

import type { Job } from './api/client';
import { subsetText, summarize } from './jobtext';

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
  it('says how much data a check read, and what to do about index errors', () => {
    expect(summarize(job('check', { num_errors: 0 }, { read_data_subset: '10%' }))).toBe('No errors found, 10% of the data read');
    expect(summarize(job('check', { num_errors: 0 }, { read_data_subset: '100%' }))).toBe('No errors found, all the data read');
    expect(summarize(job('check', { num_errors: 3, suggest_repair_index: true }))).toBe('3 errors: rebuild the index, in Settings');
  });
  it('describes the repairs and the upgrade', () => {
    expect(summarize(job('repair_snapshots', { modified: 0 }))).toBe('No snapshot needed repair');
    expect(summarize(job('repair_snapshots', { modified: 2 }))).toBe('Repaired 2 snapshots');
    expect(summarize(job('migrate', { applied: ['upgrade_repo_v2'] }))).toBe('Upgraded to format version 2');
    expect(summarize(job('migrate', { applied: [] }))).toBe('Already at the latest format');
    expect(summarize(job('repair_index', {}))).toBe('');
    expect(summarize(job('change_password', {}))).toBe('');
  });
  it('words a data subset', () => {
    expect(subsetText('')).toBe('');
    expect(subsetText(undefined)).toBe('');
    expect(subsetText('2.5%')).toBe('2.5% of the data read');
    expect(subsetText('1/5')).toBe('1/5 read');
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
