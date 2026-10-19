import { describe, expect, it } from 'vitest';

import { draftOfPlan, newPlanDraft, nextBackup, planOfSnapshot, planProblems, ready, toPlanInput } from './plan';
import { plan, repository } from './testing/fixtures';

describe('plans', () => {
  it('round-trips a plan through its draft', () => {
    const p = plan({
      tags: ['laptop', 'daily'],
      options: { compression: 'max', exclude_larger_than: '2G', one_file_system: true, exclude_caches: true, exclude_if_present: ['.nobackup'], limit_upload: 500, limit_download: 0 },
    });
    const input = toPlanInput(draftOfPlan(p));
    expect(input.name).toBe(p.name);
    expect(input.tags).toEqual(['laptop', 'daily']);
    expect(input.options).toEqual(p.options);
    expect(input.paths).toEqual(p.paths);
  });

  it('reads tags and file names from text', () => {
    const d = { ...newPlanDraft(), tags: ' a, b  c ', exclude_if_present: '.nobackup\n\n CACHEDIR.TAG \n', limit_upload: 12.7 };
    const input = toPlanInput(d);
    expect(input.tags).toEqual(['a', 'b', 'c']);
    expect(input.options!.exclude_if_present).toEqual(['.nobackup', 'CACHEDIR.TAG']);
    expect(input.options!.limit_upload).toBe(12);
  });

  it('only schedules retention when it is on', () => {
    const d = { ...newPlanDraft(), retentionOn: false };
    const input = toPlanInput(d);
    expect(input.retention).toBeNull();
    expect(input.schedules!.map((s) => s.job)).toEqual(['backup']);
  });

  it('lists what is wrong', () => {
    const d = { ...newPlanDraft('Photos'), tags: 'resticgui:x', exclude_larger_than: 'big' };
    expect(planProblems(d, ['Photos'])).toEqual([
      'Another plan is already named Photos',
      'Add at least one folder or file to back up',
      'Tags starting with resticgui: are ResticGUI’s own',
      'Skip files larger than takes a size such as 500M or 2G',
    ]);
    expect(planProblems({ ...newPlanDraft(' '), paths: { include: ['/a'], exclude: [] } })).toEqual(['Give the plan a name']);
  });

  it('finds the plans of a repository', () => {
    const photos = plan({ id: 71, name: 'Photos', tag: 'resticgui:7:abcd1234', schedules: [{ job: 'backup', cron: '0 1 * * *', enabled: true, next_run: '2026-10-19T01:00:00Z' }] });
    const empty = plan({ id: 72, name: 'Empty', tag: 'resticgui:7:00000000', paths: { include: [], exclude: [] } });
    const r = repository({ plans: [plan(), photos, empty] });
    expect(ready(r).map((p) => p.name)).toEqual(['Default', 'Photos']);
    expect(planOfSnapshot(r, ['other', 'resticgui:7:abcd1234'])?.name).toBe('Photos');
    expect(planOfSnapshot(r, ['resticgui:7'])?.name).toBe('Default');
    expect(planOfSnapshot(r, undefined)).toBeUndefined();
    expect(nextBackup(r)).toBe('2026-10-19T01:00:00Z');
  });
});
