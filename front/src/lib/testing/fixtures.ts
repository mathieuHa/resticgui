import type { Job, Repository } from '../api/client';

export const job = (over: Partial<Job>): Job => ({
  id: 1,
  repository_id: 7,
  kind: 'backup',
  trigger: 'manual',
  status: 'succeeded',
  created_at: '2026-10-18T10:00:00Z',
  started_at: '2026-10-18T10:00:00Z',
  finished_at: '2026-10-18T10:01:00Z',
  summary: {},
  params: {},
  ...over,
});

export const repository = (over: Partial<Repository> = {}): Repository => ({
  id: 7,
  name: 'Documents',
  backend: 'local',
  location: '/srv/restic/docs',
  env: {},
  secrets: [],
  has_password: true,
  extended: {},
  host: '',
  compression: 'auto',
  exclude_larger_than: '',
  initialized_at: '2026-10-01T08:00:00Z',
  created_at: '2026-10-01T08:00:00Z',
  updated_at: '2026-10-01T08:00:00Z',
  paths: { include: ['/home/me'], exclude: [] },
  retention: null,
  schedules: [{ job: 'backup', cron: '30 2 * * *', enabled: true, next_run: '2026-10-19T02:30:00Z' }],
  ...over,
});
