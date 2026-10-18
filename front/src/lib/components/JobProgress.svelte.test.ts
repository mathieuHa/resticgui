import { render, screen } from '@testing-library/svelte';
import { flushSync } from 'svelte';
import { describe, expect, it } from 'vitest';

import type { Job } from '#lib/api/client.ts';
import { jobs } from '#lib/app.svelte.ts';

import JobProgress from './JobProgress.svelte';

const job: Job = {
  id: 42,
  repository_id: 1,
  plan_id: 1,
  kind: 'backup',
  trigger: 'schedule',
  status: 'running',
  created_at: '2026-10-18T10:00:00Z',
  started_at: new Date(Date.now() - 65_000).toISOString(),
  finished_at: null,
  summary: {},
  params: {},
};

describe('JobProgress', () => {
  it('shows the progress the event stream reported', () => {
    render(JobProgress, { props: { job } });
    expect(screen.getByText('Backup')).toBeInTheDocument();
    expect(screen.getByText(/for 1m/)).toBeInTheDocument();
    // no progress yet: an indeterminate bar
    expect(document.querySelector('progress:not([value])')).not.toBeNull();

    jobs.progress[42] = {
      percent_done: 0.25,
      files_done: 10,
      total_files: 40,
      bytes_done: 1024 * 1024,
      total_bytes: 4 * 1024 * 1024,
      seconds_remaining: 90,
      current_files: ['/home/me/photo.jpg'],
    };
    flushSync();
    expect(screen.getByText('25%')).toBeInTheDocument();
    expect(screen.getByText('10 / 40 files')).toBeInTheDocument();
    expect(screen.getByText('1 MiB / 4 MiB')).toBeInTheDocument();
    expect(screen.getByText('1m 30s left')).toBeInTheDocument();
    expect(screen.getByText('/home/me/photo.jpg')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeEnabled();
  });

  it('says a queued job is waiting', () => {
    render(JobProgress, { props: { job: { ...job, id: 43, status: 'queued', started_at: null } } });
    expect(screen.getByText('waiting for the repository')).toBeInTheDocument();
  });
});
