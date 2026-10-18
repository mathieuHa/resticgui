import { flushSync } from 'svelte';
import { describe, expect, it, vi } from 'vitest';

import type { Job, JobEvent } from './api/client';
import { JobStore, type EventSourceLike } from './jobs.svelte';

class FakeSource implements EventSourceLike {
  listeners: Record<string, ((e: MessageEvent) => void)[]> = {};
  onopen: ((e: Event) => void) | null = null;
  onerror: ((e: Event) => void) | null = null;
  closed = false;
  constructor(readonly url: string) {}
  addEventListener(type: string, fn: (e: MessageEvent) => void) {
    (this.listeners[type] ??= []).push(fn);
  }
  close() {
    this.closed = true;
  }
  emit(type: string, data: Partial<JobEvent>) {
    for (const fn of this.listeners[type] ?? []) fn(new MessageEvent(type, { data: JSON.stringify(data) }));
  }
}

const job = (over: Partial<Job>): Job => ({
  id: 1,
  repository_id: 7,
  plan_id: null,
  kind: 'backup',
  trigger: 'manual',
  status: 'queued',
  created_at: '2026-10-18T10:00:00Z',
  started_at: null,
  finished_at: null,
  summary: {},
  params: {},
  ...over,
});

const NOW = new Date('2026-10-18T10:05:00Z');

function setup(fetched: Record<number, Job> = {}) {
  let source!: FakeSource;
  const fetchJob = vi.fn(async (id: number) => {
    const j = fetched[id];
    if (!j) throw new Error('not found');
    return j;
  });
  const store = new JobStore({
    fetchJob,
    createSource: (url) => (source = new FakeSource(url)),
    now: () => NOW,
  });
  store.connect('/api/v2/events');
  return { store, source, fetchJob };
}

const settle = () => new Promise((r) => setTimeout(r, 0));

describe('JobStore', () => {
  it('follows a job from queued to succeeded', async () => {
    const done = job({ status: 'succeeded', finished_at: '2026-10-18T10:04:59Z', snapshot_id: 'abcd1234', summary: { files_new: 3 } });
    const { store, source, fetchJob } = setup({ 1: done });
    store.add([job({})]);
    source.onopen?.(new Event('open'));
    expect(store.connected).toBe(true);

    source.emit('job', { job_id: 1, repository_id: 7, kind: 'backup', status: 'running' });
    await settle();
    expect(store.byId[1]?.status).toBe('running');
    expect(store.byId[1]?.started_at).toBe(NOW.toISOString());
    expect(store.active(7)?.id).toBe(1);
    expect(fetchJob).not.toHaveBeenCalled();

    source.emit('progress', { job_id: 1, repository_id: 7, progress: { percent_done: 0.5, files_done: 1, total_files: 2, bytes_done: 10, total_bytes: 20, seconds_remaining: 3 } });
    flushSync();
    expect(store.progress[1]?.percent_done).toBe(0.5);

    const changes: Job[] = [];
    store.onChange((j) => changes.push({ ...j }));
    source.emit('job', { job_id: 1, repository_id: 7, kind: 'backup', status: 'succeeded', snapshot_id: 'abcd1234' });
    await settle();
    expect(fetchJob).toHaveBeenCalledWith(1);
    expect(store.byId[1]).toEqual(done);
    expect(store.progress[1]).toBeUndefined();
    expect(store.active(7)).toBeUndefined();
    expect(changes.map((j) => j.status)).toEqual(['succeeded']);
  });

  it('loads a job it has not seen, such as one started by the scheduler', async () => {
    const scheduled = job({ id: 5, trigger: 'schedule', status: 'running' });
    const { store, source } = setup({ 5: scheduled });
    source.emit('job', { job_id: 5, repository_id: 7, kind: 'backup', status: 'running' });
    await settle();
    expect(store.latest(7)).toEqual(scheduled);
  });

  it('ignores progress that arrives after the job finished', async () => {
    const { store, source } = setup();
    store.add([job({ status: 'cancelled' })]);
    source.emit('progress', { job_id: 1, progress: { percent_done: 0.9, files_done: 0, total_files: 0, bytes_done: 0, total_bytes: 0, seconds_remaining: 0 } });
    expect(store.progress[1]).toBeUndefined();
  });

  it('lists the newest first and picks the running job as active', () => {
    const { store } = setup();
    store.add([
      job({ id: 1, status: 'succeeded' }),
      job({ id: 2, status: 'running', kind: 'check' }),
      job({ id: 3, status: 'queued' }),
      job({ id: 4, repository_id: 8 }),
    ]);
    expect(store.list(7).map((j) => j.id)).toEqual([3, 2, 1]);
    expect(store.active(7)?.id).toBe(2);
    expect(store.latest(7, 'check')?.id).toBe(2);
    expect(store.latest(9)).toBeUndefined();
  });

  it('asks pages to reload after the stream reconnects', () => {
    const { store, source } = setup();
    const resync = vi.fn();
    store.onResync(resync);
    source.onopen?.(new Event('open'));
    expect(resync).not.toHaveBeenCalled();
    source.onerror?.(new Event('error'));
    expect(store.connected).toBe(false);
    source.onopen?.(new Event('open'));
    expect(resync).toHaveBeenCalledOnce();
  });

  it('closes the stream on disconnect', () => {
    const { store, source } = setup();
    store.disconnect();
    expect(source.closed).toBe(true);
    expect(source.url).toBe('/api/v2/events');
  });
});
