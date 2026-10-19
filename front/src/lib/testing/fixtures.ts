import type { Job, Plan, Repository } from '../api/client';

export const job = (over: Partial<Job>): Job => ({
  id: 1,
  repository_id: 7,
  plan_id: over.kind && over.kind !== 'backup' && over.kind !== 'forget' ? null : 70,
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

export const plan = (over: Partial<Plan> = {}): Plan => ({
  id: 70,
  name: 'Default',
  tag: 'resticgui:7',
  tags: [],
  paths: { include: ['/home/me'], exclude: [] },
  options: {
    compression: 'auto',
    exclude_larger_than: '',
    one_file_system: false,
    exclude_caches: false,
    exclude_if_present: [],
    limit_upload: 0,
    limit_download: 0,
  },
  schedules: [{ job: 'backup', cron: '30 2 * * *', enabled: true, next_run: '2026-10-19T02:30:00Z' }],
  retention: null,
  created_at: '2026-10-01T08:00:00Z',
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
  initialized_at: '2026-10-01T08:00:00Z',
  created_at: '2026-10-01T08:00:00Z',
  updated_at: '2026-10-01T08:00:00Z',
  plans: [plan()],
  schedules: [],
  healthcheck: null,
  hooks: [],
  ...over,
});
