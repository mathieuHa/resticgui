import { describe, expect, it } from 'vitest';

import { headline, health, periodHours } from './health';
import { parse } from './schedule';
import { job, plan, repository } from './testing/fixtures';

const NOW = Date.parse('2026-10-18T12:00:00Z');
const at = (iso: string) => ({ started_at: iso, finished_at: iso });

describe('health', () => {
  it('is running while a job is queued or running', () => {
    expect(health(repository(), [job({ status: 'running', finished_at: null })], NOW)).toBe('running');
  });

  it('asks for setup until the repository exists and has a plan with paths', () => {
    expect(health(repository({ initialized_at: null }), [], NOW)).toBe('setup');
    expect(health(repository({ plans: [] }), [], NOW)).toBe('setup');
    expect(health(repository({ plans: [plan({ paths: { include: [], exclude: [] } })] }), [], NOW)).toBe('setup');
  });

  it('is never without a good backup, failed when the last one failed', () => {
    expect(health(repository(), [], NOW)).toBe('never');
    const good = job({ id: 1, ...at('2026-10-18T02:30:00Z') });
    const bad = job({ id: 2, status: 'failed', ...at('2026-10-18T03:30:00Z') });
    expect(health(repository(), [good, bad], NOW)).toBe('failed');
    expect(health(repository(), [good, job({ id: 3, status: 'interrupted', ...at('2026-10-18T04:00:00Z') })], NOW)).toBe('failed');
    // a later success clears it; a cancelled run does not count
    expect(health(repository(), [bad, job({ id: 4, ...at('2026-10-18T05:00:00Z') })], NOW)).toBe('ok');
    expect(health(repository(), [good, job({ id: 5, status: 'cancelled', ...at('2026-10-18T06:00:00Z') })], NOW)).toBe('ok');
  });

  it('is late after one and a half periods of the schedule', () => {
    const daily = repository();
    expect(health(daily, [job(at('2026-10-17T00:00:00Z'))], NOW)).toBe('ok'); // 36h ago
    expect(health(daily, [job(at('2026-10-16T09:00:00Z'))], NOW)).toBe('late'); // 51h ago
    const manual = repository({ plans: [plan({ schedules: [] })] });
    expect(health(manual, [job(at('2026-01-01T00:00:00Z'))], NOW)).toBe('ok');
    const disabled = repository({ plans: [plan({ schedules: [{ job: 'backup', cron: '@hourly', enabled: false, next_run: null }] })] });
    expect(health(disabled, [job(at('2026-10-17T00:00:00Z'))], NOW)).toBe('ok');
  });

  it('is the worst of its plans', () => {
    const photos = plan({ id: 71, name: 'Photos', tag: 'resticgui:7:aa' });
    const two = repository({ plans: [plan(), photos] });
    const good = job({ id: 1, ...at('2026-10-18T02:30:00Z') });
    // the second plan never backed up
    expect(health(two, [good], NOW)).toBe('never');
    const photosGood = job({ id: 2, plan_id: 71, ...at('2026-10-18T02:40:00Z') });
    expect(health(two, [good, photosGood], NOW)).toBe('ok');
    // a failure of one plan is not hidden by a later success of the other
    const photosBad = job({ id: 3, plan_id: 71, status: 'failed', ...at('2026-10-18T03:00:00Z') });
    const later = job({ id: 4, ...at('2026-10-18T04:00:00Z') });
    expect(health(two, [good, photosGood, photosBad, later], NOW)).toBe('failed');
    // a plan without paths waits for them without holding the others back
    const empty = repository({ plans: [plan(), plan({ id: 72, name: 'Later', paths: { include: [], exclude: [] } })] });
    expect(health(empty, [good], NOW)).toBe('ok');
  });

  it('knows the period of each preset', () => {
    expect(periodHours(parse('15 * * * *'))).toBe(1);
    expect(periodHours(parse('@every 6h'))).toBe(6);
    expect(periodHours(parse('0 3 * * 1'))).toBe(168);
    expect(periodHours(parse('0 3 1 * *'))).toBe(744);
    expect(periodHours(parse('*/5 1-3 * * *'))).toBe(168);
  });
});

describe('headline', () => {
  it('leads with the worst news', () => {
    expect(headline(['ok', 'failed', 'late'])).toEqual({ tone: 'error', text: '1 repository has a failed backup' });
    expect(headline(['ok', 'late', 'never'])).toEqual({ tone: 'warning', text: '2 repositories need your attention' });
    expect(headline(['ok', 'running'])).toMatchObject({ tone: 'info' });
    expect(headline(['ok'])).toEqual({ tone: 'success', text: 'Your repository is backed up' });
    expect(headline(['ok', 'ok', 'ok'])).toEqual({ tone: 'success', text: 'All 3 repositories are backed up' });
  });
});
