<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { api, type Repository, type Snapshot, type Stats } from '#lib/api/client.ts';
  import { attempt, jobs, notify } from '#lib/app.svelte.ts';
  import BackendIcon from '#lib/components/BackendIcon.svelte';
  import Chart from '#lib/components/Chart.svelte';
  import HealthBadge from '#lib/components/HealthBadge.svelte';
  import JobProgress from '#lib/components/JobProgress.svelte';
  import Modal from '#lib/components/Modal.svelte';
  import SnapshotBrowser from '#lib/components/SnapshotBrowser.svelte';
  import StatusStrip from '#lib/components/StatusStrip.svelte';
  import Timeline from '#lib/components/Timeline.svelte';
  import { backupPoints } from '#lib/charts.ts';
  import { bytes, dateTime, duration, relative } from '#lib/format.ts';
  import { health } from '#lib/health.ts';
  import { KIND_LABELS } from '#lib/jobtext.ts';
  import { BACKEND_LABELS } from '#lib/repository.ts';
  import { describeRetention } from '#lib/retention.ts';
  import { describe, parse } from '#lib/schedule.ts';

  import Camera from '@lucide/svelte/icons/camera';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import History from '@lucide/svelte/icons/history';
  import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
  import Play from '@lucide/svelte/icons/play';
  import SettingsIcon from '@lucide/svelte/icons/settings';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import FolderOpen from '@lucide/svelte/icons/folder-open';

  import Settings from './Settings.svelte';

  const TABS = {
    overview: { label: 'Overview', icon: LayoutDashboard },
    snapshots: { label: 'Snapshots', icon: Camera },
    runs: { label: 'Runs', icon: History },
    settings: { label: 'Settings', icon: SettingsIcon },
  } as const;
  type Tab = keyof typeof TABS;

  const id = $derived(Number(page.params.id));
  const tab = $derived((page.url.searchParams.get('tab') ?? 'overview') as Tab);

  let repo = $state<Repository | null>(null);
  let missing = $state(false);
  let snapshots = $state<Snapshot[] | null>(null);
  let snapshotsError = $state('');
  let allSnapshots = $state(false);
  let browsing = $state<Snapshot | null>(null);
  let forgetting = $state<Snapshot | null>(null);
  let forgetPrune = $state(false);
  let stats = $state<Stats | null>(null);
  let statsError = $state('');

  async function loadRepo() {
    try {
      repo = await api.repository(id);
      // enough history for the charts
      jobs.add(await api.jobs(id, 50));
    } catch {
      missing = true;
    }
  }

  async function loadSnapshots() {
    snapshotsError = '';
    try {
      snapshots = (await api.snapshots(id, allSnapshots)).reverse();
    } catch (err) {
      snapshots = [];
      snapshotsError = err instanceof Error ? err.message : String(err);
    }
  }

  async function loadStats() {
    statsError = '';
    try {
      stats = await api.stats(id);
    } catch (err) {
      statsError = err instanceof Error ? err.message : String(err);
    }
  }

  $effect(() => {
    // reload when the route's id changes
    void id;
    repo = null;
    snapshots = null;
    stats = null;
    loadRepo();
    const unsubscribe = jobs.onChange((job) => {
      if (job.repository_id !== id || job.status !== 'succeeded') return;
      if (['backup', 'forget', 'prune', 'init'].includes(job.kind)) {
        loadRepo();
        if (snapshots) loadSnapshots();
        if (stats || statsError) loadStats();
      }
    });
    const stop = jobs.onResync(loadRepo);
    return () => {
      unsubscribe();
      stop();
    };
  });

  $effect(() => {
    if (tab === 'snapshots' && snapshots === null) loadSnapshots();
    if (tab === 'overview' && stats === null && !statsError) loadStats();
  });

  const active = $derived(jobs.active(id));
  const list = $derived(jobs.list(id));
  const lastGood = $derived(list.find((j) => j.kind === 'backup' && j.status === 'succeeded'));
  const status = $derived(repo ? health(repo, list) : 'ok');
  const points = $derived(backupPoints(list, 30));

  function href(t: Tab) {
    return resolve('/repositories/[id]', { id: String(id) }) + (t === 'overview' ? '' : `?tab=${t}`);
  }

  async function start(kind: 'backup' | 'check' | 'forget' | 'prune' | 'unlock' | 'init') {
    await attempt(async () => {
      jobs.add([await api.startJob(id, kind)]);
      notify(`${KIND_LABELS[kind]} started`, 'info', 3000);
    }, `Could not start: ${KIND_LABELS[kind]}`);
  }

  async function forget() {
    if (!forgetting) return;
    const snap = forgetting;
    if (await attempt(() => api.forgetSnapshot(id, snap.id, forgetPrune), 'Could not forget the snapshot')) {
      notify(`Snapshot ${snap.short_id} forgotten`);
      forgetting = null;
      loadSnapshots();
    }
  }

  async function removed() {
    notify('Repository removed from ResticGUI. Its backups were not deleted.');
    await goto(resolve('/'));
  }
</script>

<svelte:head><title>{repo?.name ?? 'Repository'} · ResticGUI</title></svelte:head>

{#if missing}
  <div class="alert alert-warning">This repository does not exist. <a class="link" href={resolve('/')}>Back to the list</a></div>
{:else if !repo}
  <div class="flex flex-col gap-4" aria-busy="true" aria-label="Loading the repository">
    <div class="skeleton h-16 w-full"></div>
    <div class="skeleton h-10 w-80"></div>
    <div class="grid gap-4 md:grid-cols-2"><div class="skeleton h-48"></div><div class="skeleton h-48"></div></div>
  </div>
{:else}
  <a class="text-base-content/70 hover:text-base-content mb-3 inline-flex items-center gap-1 text-sm" href={resolve('/')}>
    <ArrowLeft class="size-4" aria-hidden="true" />All repositories
  </a>
  <div class="mb-4 flex flex-wrap items-center gap-3">
    <BackendIcon backend={repo.backend} size="lg" />
    <div class="min-w-0 flex-1">
      <div class="flex flex-wrap items-center gap-2">
        <h1 class="text-2xl font-bold">{repo.name}</h1>
        <HealthBadge health={status} />
      </div>
      <p class="text-base-content/70 truncate text-sm">{BACKEND_LABELS[repo.backend] ?? repo.backend} · <span class="font-mono">{repo.location}</span></p>
    </div>
    <div class="flex flex-wrap gap-2">
      <button class="btn btn-primary btn-sm" onclick={() => start('backup')} disabled={!!active || repo.paths.include.length === 0}>
        <Play class="size-4" aria-hidden="true" />Back up now
      </button>
      <div class="dropdown dropdown-end">
        <div tabindex="0" role="button" class="btn btn-sm">More<ChevronDown class="size-4" aria-hidden="true" /></div>
        <ul tabindex="-1" class="dropdown-content menu bg-base-100 rounded-box border-base-300 z-10 w-56 border p-2 shadow-lg">
          {#if !repo.initialized_at}<li><button onclick={() => start('init')}>Create in restic (init)</button></li>{/if}
          <li><button onclick={() => start('check')}>Check integrity</button></li>
          {#if repo.retention}<li><button onclick={() => start('forget')}>Apply retention now</button></li>{/if}
          <li><button onclick={() => start('prune')}>Prune unused data</button></li>
          <li><button onclick={() => start('unlock')}>Remove stale locks</button></li>
        </ul>
      </div>
    </div>
  </div>

  {#if active}
    <div class="card bg-base-100 border-base-300 animate-rise mb-4 border shadow-sm"><div class="card-body py-4"><JobProgress job={active} /></div></div>
  {/if}

  <div role="tablist" class="tabs tabs-border mb-4 overflow-x-auto">
    {#each Object.entries(TABS) as [key, t] (key)}
      <a role="tab" class="tab gap-2 {tab === key ? 'tab-active' : ''}" aria-selected={tab === key} href={href(key as Tab)}>
        <t.icon class="size-4" aria-hidden="true" />{t.label}
      </a>
    {/each}
  </div>

  {#if tab === 'overview'}
    <div class="animate-rise grid gap-4 md:grid-cols-2">
      <section class="card bg-base-100 border-base-300 border shadow-sm">
        <div class="card-body gap-3">
          <h2 class="card-title text-base">Backups</h2>
          <p class="text-lg font-medium">
            {#if lastGood}Last backup <time datetime={lastGood.finished_at} title={dateTime(lastGood.finished_at)}>{relative(lastGood.finished_at)}</time>{:else}No backup yet{/if}
          </p>
          {#if points.length}<StatusStrip {points} slots={30} />{/if}
          <ul class="text-base-content/80 flex flex-col gap-1 text-sm">
            {#each ['backup', 'forget', 'check'] as const as job (job)}
              {@const s = repo.schedules.find((x) => x.job === job)}
              <li class="flex justify-between gap-2">
                <span>{KIND_LABELS[job]}: {s?.enabled ? describe(parse(s.cron)) : 'by hand only'}</span>
                {#if s?.next_run}<span class="text-base-content/70 whitespace-nowrap">next <time datetime={s.next_run}>{relative(s.next_run)}</time></span>{/if}
              </li>
            {/each}
          </ul>
          <p class="text-base-content/80 text-sm">{describeRetention(repo.retention)}</p>
          <p class="text-base-content/80 text-sm">Backs up <span class="font-mono">{repo.paths.include.join(', ') || 'nothing yet'}</span></p>
        </div>
      </section>
      <section class="card bg-base-100 border-base-300 border shadow-sm">
        <div class="card-body">
          <h2 class="card-title text-base">Size</h2>
          {#if statsError}
            <p class="text-base-content/70 text-sm">{statsError}</p>
          {:else if !stats}
            <div class="grid grid-cols-3 gap-3"><div class="skeleton h-16"></div><div class="skeleton h-16"></div><div class="skeleton h-16"></div></div>
          {:else}
            <div class="grid grid-cols-3 gap-3">
              <div class="bg-base-200 rounded-box p-3"><div class="text-base-content/70 text-xs">Snapshots</div><div class="tabular text-2xl font-semibold">{stats.snapshots_count ?? 0}</div></div>
              <div class="bg-base-200 rounded-box p-3"><div class="text-base-content/70 text-xs">Restore size</div><div class="tabular text-2xl font-semibold">{bytes(stats.total_size)}</div></div>
              <div class="bg-base-200 rounded-box p-3"><div class="text-base-content/70 text-xs">Files</div><div class="tabular text-2xl font-semibold">{stats.total_file_count ?? 0}</div></div>
            </div>
          {/if}
        </div>
      </section>
      <section class="card bg-base-100 border-base-300 border shadow-sm md:col-span-2">
        <div class="card-body">
          <h2 class="card-title text-base">Over the last {points.length || ''} backups</h2>
          <div class="grid gap-6 sm:grid-cols-3">
            <Chart title="Data added" {points} value={(p) => p.added} format={bytes} bytes />
            <Chart title="Size backed up" {points} value={(p) => p.processed} format={bytes} bytes kind="line" tone="secondary" />
            <Chart title="Duration" {points} value={(p) => p.took} format={duration} kind="line" tone="accent" />
          </div>
        </div>
      </section>
    </div>
  {:else if tab === 'snapshots'}
    <div class="card bg-base-100 border-base-300 animate-rise border shadow-sm">
      <div class="card-body">
        {#if browsing}
          <button class="btn btn-ghost btn-sm self-start" onclick={() => (browsing = null)}><ArrowLeft class="size-4" aria-hidden="true" />All snapshots</button>
          {#key browsing.id}<SnapshotBrowser repositoryId={id} snapshot={browsing} />{/key}
        {:else}
          <label class="label cursor-pointer gap-2 self-end text-sm">
            <input type="checkbox" class="toggle toggle-xs" bind:checked={allSnapshots} onchange={loadSnapshots} />
            Also snapshots made outside ResticGUI
          </label>
          {#if snapshotsError}<div class="alert alert-warning text-sm">{snapshotsError}</div>{/if}
          {#if snapshots === null}
            <div class="flex flex-col gap-2" aria-busy="true">{#each [1, 2, 3] as i (i)}<div class="skeleton h-9"></div>{/each}</div>
          {:else if snapshots.length === 0 && !snapshotsError}
            <p class="text-base-content/70">No snapshot yet. Run a backup to make the first one.</p>
          {:else}
            <div class="overflow-x-auto">
              <table class="table-sm table" data-testid="snapshots">
                <thead><tr><th>Snapshot</th><th>Time</th><th>Host</th><th>Paths</th><th>Added</th><th></th></tr></thead>
                <tbody>
                  {#each snapshots as s (s.id)}
                    <tr class="hover:bg-base-200">
                      <td><span class="badge badge-ghost font-mono">{s.short_id}</span></td>
                      <td class="whitespace-nowrap"><time datetime={s.time}>{dateTime(s.time)}</time></td>
                      <td>{s.hostname}</td>
                      <td class="max-w-xs truncate font-mono text-xs" title={s.paths.join('\n')}>{s.paths.join(', ')}</td>
                      <td class="tabular whitespace-nowrap">{bytes((s.summary as { data_added?: number } | undefined)?.data_added)}</td>
                      <td class="whitespace-nowrap text-right">
                        <button class="btn btn-sm" onclick={() => (browsing = s)}><FolderOpen class="size-4" aria-hidden="true" />Browse</button>
                        <button class="btn btn-ghost btn-sm" onclick={() => { forgetting = s; forgetPrune = false; }}>Forget</button>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        {/if}
      </div>
    </div>
  {:else if tab === 'runs'}
    <div class="card bg-base-100 border-base-300 animate-rise border shadow-sm"><div class="card-body"><Timeline repositoryId={id} /></div></div>
  {:else}
    {#key repo.updated_at}
      <Settings {repo} onsaved={(r) => (repo = r)} onremoved={removed} />
    {/key}
  {/if}
{/if}

<Modal open={!!forgetting} title="Forget snapshot {forgetting?.short_id}" onclose={() => (forgetting = null)}>
  <p class="mb-3">The snapshot from {dateTime(forgetting?.time)} is removed from the repository. This cannot be undone.</p>
  <label class="label cursor-pointer gap-2">
    <input type="checkbox" class="checkbox checkbox-sm" bind:checked={forgetPrune} />
    <span class="text-base-content">Also prune, to free its space now</span>
  </label>
  {#snippet actions()}
    <button class="btn" onclick={() => (forgetting = null)}>Cancel</button>
    <button class="btn btn-error" onclick={forget}>Forget</button>
  {/snippet}
</Modal>
