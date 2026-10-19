<script lang="ts">
  import { untrack } from 'svelte';

  import { api, type Backend, type Repository } from '#lib/api/client.ts';
  import { attempt, jobs, notify } from '#lib/app.svelte.ts';
  import Modal from '#lib/components/Modal.svelte';
  import RepositoryForm from '#lib/components/RepositoryForm.svelte';
  import ScheduleEditor from '#lib/components/ScheduleEditor.svelte';
  import Maintenance from './Maintenance.svelte';
  import Monitoring from './Monitoring.svelte';
  import PlanSettings from './PlanSettings.svelte';
  import { SUBSETS } from '#lib/jobtext.ts';
  import { draftOf, problems, toInput } from '#lib/repository.ts';

  import Plus from '@lucide/svelte/icons/plus';

  let { repo, onsaved, onremoved }: { repo: Repository; onsaved: (r: Repository) => void; onremoved: () => void } = $props();

  // each section starts from the repository as it was when the tab opened
  const initial = untrack(() => repo);
  let backends = $state<Backend[]>([]);
  let draft = $state(draftOf(initial));
  const check = initial.schedules.find((x) => x.job === 'check');
  let checkCron = $state(check?.enabled ? check.cron : null);
  let checkSubset = $state(check?.read_data_subset ?? '');
  let adding = $state(false);
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

  async function saveCheck() {
    await attempt(async () => {
      if (checkCron) onsaved(await api.setCheckSchedule(repo.id, checkCron, checkSubset));
      else if (repo.schedules.some((s) => s.job === 'check')) onsaved(await api.deleteCheckSchedule(repo.id));
      notify('Integrity check saved');
    }, 'Could not save the integrity check');
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

  <section class="card bg-base-100 border-base-300 border shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">Storage</h2>
      <RepositoryForm bind:draft {backends} existing={repo} />
      <div class="card-actions justify-end"><button class="btn btn-primary btn-sm" onclick={saveStorage}>Save storage</button></div>
    </div>
  </section>

  <div class="flex flex-col gap-3" data-testid="plans">
    <div class="flex flex-wrap items-end justify-between gap-2">
      <div>
        <h2 class="text-lg font-semibold">Backup plans</h2>
        <p class="text-base-content/70 text-sm">Each plan backs up its own folders on its own schedule, all into this repository, where identical data is stored once. Times are in the server's time zone.</p>
      </div>
      {#if !adding}
        <button class="btn btn-sm" onclick={() => (adding = true)}><Plus class="size-4" aria-hidden="true" />Add a plan</button>
      {/if}
    </div>
    {#each repo.plans as plan (plan.id)}
      <PlanSettings {repo} {plan} {busy} {onsaved} />
    {/each}
    {#if adding}
      <PlanSettings {repo} onsaved={(r) => { adding = false; onsaved(r); }} oncancel={() => (adding = false)} />
    {:else if !repo.plans.length}
      <p class="text-base-content/70 text-sm">No plan yet: add one to choose what to back up.</p>
    {/if}
  </div>

  <section class="card bg-base-100 border-base-300 border shadow-sm">
    <div class="card-body">
      <h2 class="card-title text-base">Integrity check</h2>
      <ScheduleEditor label="Check the repository's integrity" bind:value={checkCron} />
      {#if checkCron}
        <label class="flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-2">
          <span class="text-sm">Also read</span>
          <select class="select select-sm" bind:value={checkSubset} aria-label="Data the scheduled check reads">
            {#each SUBSETS as [value, label] (value)}<option {value}>{label}</option>{/each}
          </select>
        </label>
        <p class="text-base-content/70 text-xs">Reading data finds damaged files, but downloads them from the storage. Each check reads another random part.</p>
      {/if}
      <div class="card-actions justify-end"><button class="btn btn-primary btn-sm" onclick={saveCheck}>Save check</button></div>
    </div>
  </section>

  <Maintenance {repo} />

  <Monitoring {repo} {onsaved} />

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
