<script lang="ts">
  import { untrack } from 'svelte';

  import { api, type Backend, type Repository, type Retention, type ScheduledJob } from '#lib/api/client.ts';
  import { attempt, jobs, notify } from '#lib/app.svelte.ts';
  import Modal from '#lib/components/Modal.svelte';
  import PathsEditor from '#lib/components/PathsEditor.svelte';
  import RepositoryForm from '#lib/components/RepositoryForm.svelte';
  import RetentionEditor from '#lib/components/RetentionEditor.svelte';
  import ScheduleEditor from '#lib/components/ScheduleEditor.svelte';
  import { draftOf, problems, toInput } from '#lib/repository.ts';
  import { isEmpty, normalize, suggested } from '#lib/retention.ts';

  let { repo, onsaved, onremoved }: { repo: Repository; onsaved: (r: Repository) => void; onremoved: () => void } = $props();

  // each section starts from the repository as it was when the tab opened
  const initial = untrack(() => repo);
  let backends = $state<Backend[]>([]);
  let draft = $state(draftOf(initial));
  let paths = $state($state.snapshot(initial.paths));
  const cronOf = (job: ScheduledJob) => {
    const s = initial.schedules.find((x) => x.job === job);
    return s?.enabled ? s.cron : null;
  };
  let schedules: Record<ScheduledJob, string | null> = $state({
    backup: cronOf('backup'),
    forget: cronOf('forget'),
    check: cronOf('check'),
  });
  let retentionOn = $state(!!initial.retention);
  let retention: Retention = $state(initial.retention ? { ...initial.retention } : suggested());
  let confirmDelete = $state(false);
  let errors = $state<string[]>([]);

  $effect(() => {
    api.backends().then((b) => (backends = b));
  });
  const backend = $derived(backends.find((b) => b.name === draft.backend));

  async function saveStorage() {
    errors = problems(draft, backend, false);
    if (errors.length) return;
    await attempt(async () => {
      onsaved(await api.updateRepository(repo.id, toInput(draft, backend)));
      notify('Storage settings saved');
    }, 'Could not save the storage settings');
  }

  async function savePaths() {
    await attempt(async () => {
      onsaved(await api.setPaths(repo.id, $state.snapshot(paths)));
      notify('Paths saved');
    }, 'Could not save the paths');
  }

  async function saveSchedules() {
    await attempt(async () => {
      let latest = repo;
      for (const job of ['backup', 'forget', 'check'] as const) {
        const cron = schedules[job];
        const had = repo.schedules.some((s) => s.job === job);
        if (cron) latest = await api.setSchedule(repo.id, job, cron);
        else if (had) latest = await api.deleteSchedule(repo.id, job);
      }
      onsaved(latest);
      notify('Schedules saved');
    }, 'Could not save the schedules');
  }

  async function saveRetention() {
    if (retentionOn && isEmpty(retention)) {
      errors = ['Set at least one retention rule, or turn retention off'];
      return;
    }
    errors = [];
    await attempt(async () => {
      if (retentionOn) onsaved(await api.setRetention(repo.id, normalize($state.snapshot(retention))));
      else if (repo.retention) onsaved(await api.deleteRetention(repo.id));
      notify('Retention saved');
    }, 'Could not save the retention policy');
  }

  async function remove() {
    if (await attempt(() => api.deleteRepository(repo.id), 'Could not remove the repository')) onremoved();
  }

  const busy = $derived(!!jobs.active(repo.id));
</script>

<div class="flex flex-col gap-4">
  {#if errors.length}
    <ul class="alert alert-error flex-col items-start text-sm">
      {#each errors as e (e)}<li>{e}</li>{/each}
    </ul>
  {/if}

  <section class="card bg-base-100 shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">Storage</h2>
      <RepositoryForm bind:draft {backends} existing={repo} />
      <div class="card-actions justify-end"><button class="btn btn-primary btn-sm" onclick={saveStorage}>Save storage</button></div>
    </div>
  </section>

  <section class="card bg-base-100 shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">What to back up</h2>
      <PathsEditor bind:value={paths} />
      <div class="card-actions justify-end"><button class="btn btn-primary btn-sm" onclick={savePaths}>Save paths</button></div>
    </div>
  </section>

  <section class="card bg-base-100 shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">Schedules</h2>
      <p class="text-xs opacity-60">Times are in the server's time zone.</p>
      <ScheduleEditor label="Back up automatically" bind:value={schedules.backup} />
      <ScheduleEditor label="Apply the retention policy automatically" bind:value={schedules.forget} />
      <ScheduleEditor label="Check the repository's integrity" bind:value={schedules.check} />
      <div class="card-actions justify-end"><button class="btn btn-primary btn-sm" onclick={saveSchedules}>Save schedules</button></div>
    </div>
  </section>

  <section class="card bg-base-100 shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">Retention</h2>
      <label class="label cursor-pointer gap-3">
        <input type="checkbox" class="toggle toggle-sm" bind:checked={retentionOn} />
        <span class="text-base-content">Remove old snapshots</span>
      </label>
      {#if retentionOn}<RetentionEditor bind:value={retention} repositoryId={repo.id} />{/if}
      <div class="card-actions justify-end"><button class="btn btn-primary btn-sm" onclick={saveRetention}>Save retention</button></div>
    </div>
  </section>

  <section class="card bg-base-100 border-error/30 border shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">Remove from ResticGUI</h2>
      <p class="text-sm">ResticGUI forgets this repository and its history of runs. The backups stay in the repository.</p>
      <div class="card-actions justify-end">
        <button class="btn btn-error btn-sm" onclick={() => (confirmDelete = true)} disabled={busy} title={busy ? 'Wait for the running job to finish' : ''}>Remove</button>
      </div>
    </div>
  </section>
</div>

<Modal bind:open={confirmDelete} title="Remove {repo.name}?">
  <p>The repository's backups are not deleted, and you can add it again later with its password.</p>
  {#snippet actions()}
    <button class="btn" onclick={() => (confirmDelete = false)}>Cancel</button>
    <button class="btn btn-error" onclick={remove}>Remove</button>
  {/snippet}
</Modal>
