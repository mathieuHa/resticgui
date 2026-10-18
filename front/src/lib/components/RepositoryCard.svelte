<script lang="ts">
  import { resolve } from '$app/paths';
  import { api, type Repository } from '#lib/api/client.ts';
  import { attempt, jobs } from '#lib/app.svelte.ts';
  import { relative } from '#lib/format.ts';
  import { BACKEND_LABELS } from '#lib/repository.ts';
  import { describe, parse } from '#lib/schedule.ts';

  import JobProgress from './JobProgress.svelte';

  let { repository: r }: { repository: Repository } = $props();

  const active = $derived(jobs.active(r.id));
  const lastBackup = $derived(jobs.latest(r.id, 'backup'));
  const lastGood = $derived(jobs.list(r.id).find((j) => j.kind === 'backup' && j.status === 'succeeded'));
  const backupSchedule = $derived(r.schedules.find((s) => s.job === 'backup' && s.enabled));

  async function backup() {
    await attempt(async () => jobs.add([await api.startJob(r.id, 'backup')]), 'Could not start the backup');
  }
  async function init() {
    await attempt(async () => jobs.add([await api.startJob(r.id, 'init')]), 'Could not create the repository');
  }
</script>

<div class="card bg-base-100 shadow-sm" data-testid="repository-card">
  <div class="card-body gap-3">
    <div class="flex items-start gap-2">
      <div class="min-w-0 flex-1">
        <h2 class="card-title"><a class="link link-hover" href={resolve('/repositories/[id]', { id: String(r.id) })}>{r.name}</a></h2>
        <p class="truncate text-sm opacity-60" title={r.location}>{BACKEND_LABELS[r.backend] ?? r.backend} · <span class="font-mono">{r.location}</span></p>
      </div>
    </div>

    {#if active}
      <JobProgress job={active} />
    {:else if lastBackup?.status === 'failed'}
      <p class="text-error text-sm">Last backup failed {relative(lastBackup.finished_at)}: {lastBackup.error}</p>
    {:else}
      <p class="text-sm">
        {lastGood ? `Last backup ${relative(lastGood.finished_at)}` : 'No backup yet'}
      </p>
    {/if}
    <p class="text-xs opacity-60">
      {backupSchedule ? describe(parse(backupSchedule.cron)) : 'No backup schedule'}
      · {r.paths.include.length} {r.paths.include.length === 1 ? 'path' : 'paths'}
    </p>

    <div class="card-actions flex-wrap justify-end">
      {#if !r.initialized_at}
        <button class="btn btn-sm" onclick={init} disabled={!!active} title="Run restic init to create a new repository at this location">Create in restic</button>
      {/if}
      <a class="btn btn-sm" href={resolve('/repositories/[id]', { id: String(r.id) })}>Open</a>
      <button class="btn btn-primary btn-sm" onclick={backup} disabled={!!active || r.paths.include.length === 0} title={r.paths.include.length === 0 ? 'Choose what to back up in Settings' : ''}>Back up now</button>
    </div>
  </div>
</div>
