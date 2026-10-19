<script lang="ts">
  import Calendar from '@lucide/svelte/icons/calendar';
  import Files from '@lucide/svelte/icons/file';
  import Play from '@lucide/svelte/icons/play';
  import { resolve } from '$app/paths';
  import { untrack } from 'svelte';

  import { api, type Repository } from '#lib/api/client.ts';
  import { attempt, jobs } from '#lib/app.svelte.ts';
  import { backupPoints } from '#lib/charts.ts';
  import { bytes, relative } from '#lib/format.ts';
  import { health } from '#lib/health.ts';
  import { nextBackup, ready } from '#lib/plan.ts';
  import { BACKEND_LABELS } from '#lib/repository.ts';

  import BackendIcon from './BackendIcon.svelte';
  import Checklist from './Checklist.svelte';
  import HealthBadge from './HealthBadge.svelte';
  import JobProgress from './JobProgress.svelte';
  import StatusStrip from './StatusStrip.svelte';

  let { repository: r, now = Date.now() }: { repository: Repository; now?: number } = $props();

  const list = $derived(jobs.list(r.id));
  const active = $derived(jobs.active(r.id));
  const status = $derived(health(r, list, now));
  const lastBackup = $derived(list.find((j) => j.kind === 'backup' && j.finished_at && j.status !== 'cancelled'));
  const lastGood = $derived(list.find((j) => j.kind === 'backup' && j.status === 'succeeded'));
  const plans = $derived(ready(r));
  const next = $derived(nextBackup(r));
  const points = $derived(backupPoints(list, 20));
  const summary = $derived((lastGood?.summary ?? {}) as { total_bytes_processed?: number; total_files_processed?: number });
  const href = $derived(resolve('/repositories/[id]', { id: String(r.id) }));

  // a short glow when a backup finishes while the page is open
  let glow = $state(false);
  let seen = untrack(() => lastGood?.id);
  $effect(() => {
    const id = lastGood?.id;
    if (id === undefined || id === seen) return;
    seen = id;
    glow = true;
    const t = setTimeout(() => (glow = false), 1000);
    return () => clearTimeout(t);
  });

  // backs up every plan that has something to back up
  async function backup() {
    await attempt(async () => {
      for (const p of plans) jobs.add([await api.startJob(r.id, 'backup', p.id)]);
    }, 'Could not start the backup');
  }
  async function init() {
    await attempt(async () => jobs.add([await api.startJob(r.id, 'init')]), 'Could not create the repository');
  }
</script>

<article
  class="card bg-base-100 border-base-300 min-w-0 border shadow-sm transition-shadow hover:shadow-md {glow ? 'animate-done' : ''}
    {status === 'failed' ? 'border-l-error border-l-4' : status === 'late' ? 'border-l-warning border-l-4' : ''}"
  data-testid="repository-card"
>
  <div class="card-body gap-4 p-5">
    <header class="flex items-start gap-3">
      <BackendIcon backend={r.backend} />
      <div class="min-w-0 flex-1">
        <h2 class="truncate text-base font-semibold"><a class="link link-hover" {href}>{r.name}</a></h2>
        <p class="text-base-content/70 truncate text-xs" title={r.location}>{BACKEND_LABELS[r.backend] ?? r.backend} · <span class="font-mono">{r.location}</span></p>
      </div>
      <HealthBadge health={status} />
    </header>

    {#if active}
      <JobProgress job={active} />
    {:else if status === 'setup' || status === 'never'}
      <Checklist
        steps={[
          { title: 'Create the repository in restic', done: !!r.initialized_at, action: initAction },
          { title: 'Choose what to back up', done: plans.length > 0, action: pathsAction },
          { title: 'Run the first backup', done: !!lastGood, action: backupAction },
        ]}
      />
      {#if lastBackup?.status === 'failed'}<p class="text-error text-sm">The last attempt failed: {lastBackup.error}</p>{/if}
    {:else}
      <div>
        {#if status === 'failed' && lastBackup}
          <p class="text-error font-medium">Last backup failed <time datetime={lastBackup.finished_at}>{relative(lastBackup.finished_at, now)}</time></p>
          <p class="text-error/90 line-clamp-2 text-sm">{lastBackup.error}</p>
        {:else}
          <p class="text-lg font-medium">Last backup <time datetime={lastGood?.finished_at}>{relative(lastGood?.finished_at, now)}</time></p>
        {/if}
        <div class="text-base-content/70 mt-1 flex flex-wrap gap-x-4 gap-y-1 text-sm">
          <span class="flex items-center gap-1.5">
            <Calendar class="size-4" aria-hidden="true" />
            {#if next}Next <time datetime={next}>{relative(next, now)}</time>{:else}By hand only{/if}
          </span>
          {#if summary.total_bytes_processed !== undefined}
            <span class="flex items-center gap-1.5">
              <Files class="size-4" aria-hidden="true" />
              {bytes(summary.total_bytes_processed)} in {summary.total_files_processed ?? 0} files
            </span>
          {/if}
        </div>
      </div>
    {/if}

    {#if points.length}
      <StatusStrip {points} />
    {/if}

    <footer class="card-actions items-center justify-end">
      <a class="btn btn-ghost btn-sm" {href}>Open</a>
      <button
        class="btn btn-primary btn-sm"
        onclick={backup}
        disabled={!!active || plans.length === 0 || !r.initialized_at}
        title={plans.length === 0 ? 'Choose what to back up in Settings' : plans.length > 1 ? `Backs up its ${plans.length} plans` : ''}
      >
        <Play class="size-4" aria-hidden="true" />Back up now
      </button>
    </footer>
  </div>
</article>

{#snippet initAction()}
  <button class="btn btn-xs btn-primary" onclick={init} disabled={!!active} title="Run restic init to create a new repository at this location">Create in restic</button>
{/snippet}
{#snippet pathsAction()}
  <a class="btn btn-xs btn-primary" href="{href}?tab=settings">Choose</a>
{/snippet}
{#snippet backupAction()}
  <span class="text-base-content/70 text-xs">Use Back up now</span>
{/snippet}
