import { describe, expect, it } from 'vitest';

import { backupPoints, bars, linePath, niceMax } from './charts';
import { job } from './testing/fixtures';

const backup = (id: number, finished: string, summary: Record<string, unknown> = {}, status: 'succeeded' | 'failed' = 'succeeded') =>
  job({ id, status, started_at: finished, finished_at: finished, summary });

describe('backupPoints', () => {
  it('keeps finished backups, oldest first, and the newest at most', () => {
    const jobs = [
      backup(3, '2026-10-18T03:00:00Z', { data_added: 30, total_bytes_processed: 300, total_duration: 3, total_files_processed: 3 }),
      backup(1, '2026-10-16T03:00:00Z', { data_added: 10 }),
      backup(2, '2026-10-17T03:00:00Z', {}, 'failed'),
      job({ id: 4, kind: 'check' }),
      job({ id: 5, status: 'running', finished_at: null }),
    ];
    const points = backupPoints(jobs);
    expect(points.map((p) => p.id)).toEqual([1, 2, 3]);
    expect(points[2]).toMatchObject({ added: 30, processed: 300, took: 3, files: 3 });
    expect(points[1]!.added).toBeUndefined();
    expect(backupPoints(jobs, 2).map((p) => p.id)).toEqual([2, 3]);
  });

  it('falls back to the job times for the duration', () => {
    const p = backupPoints([job({ started_at: '2026-10-18T10:00:00Z', finished_at: '2026-10-18T10:02:30Z' })]);
    expect(p[0]!.took).toBe(150);
  });
});

describe('niceMax', () => {
  it('rounds up to 1, 2, 2.5 or 5 times a power', () => {
    expect([0, 0.3, 1, 7, 12, 21, 26, 480, 999].map((n) => niceMax(n))).toEqual([1, 0.5, 1, 10, 20, 25, 50, 500, 1000]);
  });
  it('rounds bytes within their unit', () => {
    expect(niceMax(700 * 1024 * 1024, 1024)).toBe(1000 * 1024 * 1024);
    expect(niceMax(3 * 1024 ** 3, 1024)).toBe(5 * 1024 ** 3);
  });
});

describe('bars and lines', () => {
  const points = backupPoints([
    backup(1, '2026-10-16T03:00:00Z', { data_added: 50 }),
    backup(2, '2026-10-17T03:00:00Z', {}, 'failed'),
    backup(3, '2026-10-18T03:00:00Z', { data_added: 100 }),
  ]);
  const added = (p: (typeof points)[number]) => p.added;

  it('scales bars to the box and leaves a gap for a failed run', () => {
    const b = bars(points, added, 300, 100, 100);
    expect(b.map((x) => x.height)).toEqual([50, 0, 100]);
    expect(b.map((x) => x.y)).toEqual([50, 100, 0]);
    expect(b[1]!.x).toBeCloseTo(100 + b[0]!.x);
  });

  it('gives a tiny value a visible bar', () => {
    expect(bars(points, (p) => (p.added ? 0.001 : undefined), 300, 100, 100)[0]!.height).toBe(1);
  });

  it('breaks the line where a value is missing', () => {
    expect(linePath(points, added, 200, 100, 100)).toBe('M0,50M200,0');
    expect(linePath(points.slice(0, 1), added, 200, 100, 100)).toBe('M100,50');
    expect(linePath([], added, 200, 100, 100)).toBe('');
  });
});
