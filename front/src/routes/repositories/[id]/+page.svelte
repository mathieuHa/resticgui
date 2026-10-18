<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { api, type Repository, type Snapshot, type Stats } from '#lib/api/client.ts';
  import { attempt, jobs, notify } from '#lib/app.svelte.ts';
  import JobProgress from '#lib/components/JobProgress.svelte';
  import RunHistory from '#lib/components/RunHistory.svelte';
  import SnapshotBrowser from '#lib/components/SnapshotBrowser.svelte';
  import Modal from '#lib/components/Modal.svelte';
  import { bytes, dateTime, relative } from '#lib/format.ts';
  import { KIND_LABELS } from '#lib/jobtext.ts';
  import { BACKEND_LABELS } from '#lib/repository.ts';
  import { describeRetention } from '#lib/retention.ts';
  import { describe, parse } from '#lib/schedule.ts';

  import Settings from './Settings.svelte';

  const TABS = { overview: 'Overview', snapshots: 'Snapshots', runs: 'Runs', settings: 'Settings' } as const;
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
      jobs.add(await api.jobs(id, 20));
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
  const lastGood = $derived(jobs.list(id).find((j) => j.kind === 'backup' && j.status === 'succeeded'));

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
  <span class="loading loading-spinner"></span>
{:else}
  <div class="mb-4 flex flex-wrap items-start gap-3">
    <div class="min-w-0 flex-1">
      <h1 class="text-2xl font-bold">{repo.name}</h1>
      <p class="truncate text-sm opacity-60">{BACKEND_LABELS[repo.backend] ?? repo.backend} · <span class="font-mono">{repo.location}</span></p>
    </div>
    <div class="flex flex-wrap gap-2">
      <button class="btn btn-primary btn-sm" onclick={() => start('backup')} disabled={!!active || repo.paths.include.length === 0}>Back up now</button>
      <div class="dropdown dropdown-end">
        <div tabindex="0" role="button" class="btn btn-sm">More</div>
        <ul tabindex="-1" class="dropdown-content menu bg-base-100 rounded-box z-10 w-56 p-2 shadow">
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
    <div class="card bg-base-100 mb-4 shadow-sm"><div class="card-body py-4"><JobProgress job={active} /></div></div>
  {/if}

  <div role="tablist" class="tabs tabs-border mb-4">
    {#each Object.entries(TABS) as [key, label] (key)}
      <a role="tab" class="tab {tab === key ? 'tab-active' : ''}" href={href(key as Tab)}>{label}</a>
    {/each}
  </div>

  {#if tab === 'overview'}
    <div class="grid gap-4 md:grid-cols-2">
      <div class="card bg-base-100 shadow-sm">
        <div class="card-body">
          <h2 class="card-title text-base">Backups</h2>
          <p>{lastGood ? `Last backup ${relative(lastGood.finished_at)} (${dateTime(lastGood.finished_at)})` : 'No backup yet'}</p>
          <ul class="text-sm opacity-80">
            {#each ['backup', 'forget', 'check'] as const as job (job)}
              {@const s = repo.schedules.find((x) => x.job === job)}
              <li>{KIND_LABELS[job]}: {s?.enabled ? describe(parse(s.cron)) : 'by hand only'}</li>
            {/each}
          </ul>
          <p class="text-sm opacity-80">{describeRetention(repo.retention)}</p>
          <p class="text-sm opacity-80">Backs up <span class="font-mono">{repo.paths.include.join(', ') || 'nothing yet'}</span></p>
        </div>
      </div>
      <div class="card bg-base-100 shadow-sm">
        <div class="card-body">
          <h2 class="card-title text-base">Size</h2>
          {#if statsError}
            <p class="text-sm opacity-70">{statsError}</p>
          {:else if !stats}
            <span class="loading loading-spinner loading-sm"></span>
          {:else}
            <div class="stats stats-vertical sm:stats-horizontal">
              <div class="stat px-2"><div class="stat-title">Snapshots</div><div class="stat-value text-2xl">{stats.snapshots_count ?? 0}</div></div>
              <div class="stat px-2"><div class="stat-title">Restore size</div><div class="stat-value text-2xl">{bytes(stats.total_size)}</div></div>
              <div class="stat px-2"><div class="stat-title">Files</div><div class="stat-value text-2xl">{stats.total_file_count ?? 0}</div></div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {:else if tab === 'snapshots'}
    <div class="card bg-base-100 shadow-sm">
      <div class="card-body">
        {#if browsing}
          <button class="btn btn-ghost btn-sm self-start" onclick={() => (browsing = null)}>← All snapshots</button>
          {#key browsing.id}<SnapshotBrowser repositoryId={id} snapshot={browsing} />{/key}
        {:else}
          <label class="label cursor-pointer gap-2 self-end text-sm">
            <input type="checkbox" class="toggle toggle-xs" bind:checked={allSnapshots} onchange={loadSnapshots} />
            Also snapshots made outside ResticGUI
          </label>
          {#if snapshotsError}<div class="alert alert-warning text-sm">{snapshotsError}</div>{/if}
          {#if snapshots === null}
            <span class="loading loading-spinner"></span>
          {:else if snapshots.length === 0 && !snapshotsError}
            <p class="opacity-60">No snapshot yet.</p>
          {:else}
            <div class="overflow-x-auto">
              <table class="table-sm table" data-testid="snapshots">
                <thead><tr><th>Snapshot</th><th>Time</th><th>Host</th><th>Paths</th><th>Added</th><th></th></tr></thead>
                <tbody>
                  {#each snapshots as s (s.id)}
                    <tr>
                      <td class="font-mono">{s.short_id}</td>
                      <td class="whitespace-nowrap">{dateTime(s.time)}</td>
                      <td>{s.hostname}</td>
                      <td class="max-w-xs truncate font-mono text-xs" title={s.paths.join('\n')}>{s.paths.join(', ')}</td>
                      <td class="whitespace-nowrap">{bytes((s.summary as { data_added?: number } | undefined)?.data_added)}</td>
                      <td class="whitespace-nowrap text-right">
                        <button class="btn btn-sm" onclick={() => (browsing = s)}>Browse</button>
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
    <div class="card bg-base-100 shadow-sm"><div class="card-body"><RunHistory repositoryId={id} /></div></div>
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
