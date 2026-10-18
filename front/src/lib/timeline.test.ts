import { describe, expect, it } from 'vitest';

import { byDay } from './timeline';
import { job } from './testing/fixtures';

describe('byDay', () => {
  it('groups jobs by local day, newest first, naming today and yesterday', () => {
    const now = new Date(2026, 9, 18, 15, 0);
    const local = (d: number, h: number) => new Date(2026, 9, d, h, 0).toISOString();
    const days = byDay(
      [
        job({ id: 1, started_at: local(16, 2) }),
        job({ id: 2, started_at: local(18, 2) }),
        job({ id: 3, started_at: local(17, 23) }),
        job({ id: 4, started_at: local(18, 9) }),
        job({ id: 5, started_at: null, created_at: local(18, 14), status: 'queued' }),
      ],
      now,
    );
    expect(days.map((d) => d.label).slice(0, 2)).toEqual(['Today', 'Yesterday']);
    // older days are written out in the browser's locale
    expect(days[2]!.label).toContain('16');
    expect(days[0]!.jobs.map((j) => j.id)).toEqual([5, 4, 2]);
    expect(days[1]!.jobs.map((j) => j.id)).toEqual([3]);
  });
});
