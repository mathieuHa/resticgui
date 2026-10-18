<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { api, type Backend, type Paths, type Retention } from '#lib/api/client.ts';
  import { jobs, notify, notifyError } from '#lib/app.svelte.ts';
  import PathsEditor from '#lib/components/PathsEditor.svelte';
  import RepositoryForm from '#lib/components/RepositoryForm.svelte';
  import RetentionEditor from '#lib/components/RetentionEditor.svelte';
  import ScheduleEditor from '#lib/components/ScheduleEditor.svelte';
  import { BACKEND_LABELS, newDraft, problems, toInput } from '#lib/repository.ts';
  import { describeRetention, isEmpty, normalize, suggested } from '#lib/retention.ts';
  import { describe, parse } from '#lib/schedule.ts';

  const STEPS = ['Storage', 'What to back up', 'When', 'Review'];
  let step = $state(0);
  let backends = $state<Backend[]>([]);
  let draft = $state(newDraft());
  let isNew = $state(true);
  let paths: Paths = $state({ include: [], exclude: [] });
  let backupCron: string | null = $state('30 2 * * *');
  let retentionOn = $state(true);
  let retention: Retention = $state(suggested());
  let forgetCron: string | null = $state('0 4 * * 0');
  let checkCron: string | null = $state(null);
  let saving = $state(false);
  let shown = $state<string[]>([]);

  $effect(() => {
    api.backends().then(
      (b) => (backends = b),
      (err) => notifyError(err, 'Could not load the storage types'),
    );
  });

  const backend = $derived(backends.find((b) => b.name === draft.backend));
  const stepProblems = $derived.by(() => {
    if (step === 0) return problems(draft, backend, true);
    if (step === 1) return paths.include.length ? [] : ['Add at least one folder or file to back up'];
    if (step === 2 && retentionOn && isEmpty(retention)) return ['Set at least one retention rule, or turn retention off'];
    return [];
  });

  function next() {
    shown = stepProblems;
    if (shown.length === 0) step++;
  }

  async function create() {
    saving = true;
    let id: number | null = null;
    try {
      const repo = await api.createRepository(toInput(draft, backend));
      id = repo.id;
      await api.setPaths(id, $state.snapshot(paths));
      if (backupCron) await api.setSchedule(id, 'backup', backupCron);
      if (retentionOn) {
        await api.setRetention(id, normalize($state.snapshot(retention)));
        if (forgetCron) await api.setSchedule(id, 'forget', forgetCron);
      }
      if (checkCron) await api.setSchedule(id, 'check', checkCron);
      if (isNew) jobs.add([await api.startJob(id, 'init')]);
      notify(isNew ? `${repo.name} is being created` : `${repo.name} added`);
      await goto(resolve('/repositories/[id]', { id: String(id) }));
    } catch (err) {
      if (id === null) {
        notifyError(err, 'Could not add the repository');
        step = 0;
      } else {
        notifyError(err, 'The repository was added, but not all its settings were saved. Check them under Settings.');
        await goto(resolve(`/repositories/[id]?tab=settings`, { id: String(id) }));
      }
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head><title>New repository · ResticGUI</title></svelte:head>

<div class="card bg-base-100 mx-auto max-w-3xl shadow-sm">
  <div class="card-body gap-6">
    <h1 class="card-title">New repository</h1>
    <ul class="steps w-full">
      {#each STEPS as name, i (name)}
        <li class="step {i <= step ? 'step-primary' : ''}">{name}</li>
      {/each}
    </ul>

    {#if step === 0}
      <div class="join">
        <button class="btn join-item btn-sm {isNew ? 'btn-active' : ''}" onclick={() => (isNew = true)}>Create a new repository</button>
        <button class="btn join-item btn-sm {!isNew ? 'btn-active' : ''}" onclick={() => (isNew = false)}>Use an existing one</button>
      </div>
      <RepositoryForm bind:draft {backends} />
    {:else if step === 1}
      <PathsEditor bind:value={paths} />
    {:else if step === 2}
      <ScheduleEditor label="Back up automatically" bind:value={backupCron} />
      <div class="divider my-0"></div>
      <label class="label cursor-pointer gap-3">
        <input type="checkbox" class="toggle toggle-sm" bind:checked={retentionOn} />
        <span class="text-base-content font-medium">Remove old snapshots</span>
      </label>
      {#if retentionOn}
        <RetentionEditor bind:value={retention} />
        <ScheduleEditor label="Apply the retention policy automatically" bind:value={forgetCron} />
      {/if}
      <div class="divider my-0"></div>
      <ScheduleEditor label="Check the repository's integrity" bind:value={checkCron} />
    {:else}
      <dl class="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
        <dt class="opacity-60">Name</dt><dd>{draft.name}</dd>
        <dt class="opacity-60">Storage</dt><dd>{BACKEND_LABELS[draft.backend]} · <span class="font-mono">{draft.location}</span></dd>
        <dt class="opacity-60">Repository</dt><dd>{isNew ? 'Created now with restic init' : 'Already exists'}</dd>
        <dt class="opacity-60">Backs up</dt><dd class="font-mono">{paths.include.join(', ')}</dd>
        {#if paths.exclude.length}<dt class="opacity-60">Excludes</dt><dd class="font-mono">{paths.exclude.join(', ')}</dd>{/if}
        <dt class="opacity-60">Backups</dt><dd>{backupCron ? describe(parse(backupCron)) : 'Only when started by hand'}</dd>
        <dt class="opacity-60">Retention</dt>
        <dd>{retentionOn ? describeRetention(retention) : 'Keep everything'}{retentionOn && forgetCron ? `, ${describe(parse(forgetCron)).toLowerCase()}` : ''}</dd>
        <dt class="opacity-60">Check</dt><dd>{checkCron ? describe(parse(checkCron)) : 'Only when started by hand'}</dd>
      </dl>
    {/if}

    {#if shown.length}
      <ul class="text-error text-sm" data-testid="problems">
        {#each shown as p (p)}<li>{p}</li>{/each}
      </ul>
    {/if}

    <div class="card-actions justify-between">
      {#if step > 0}
        <button class="btn" onclick={() => { shown = []; step--; }}>Back</button>
      {:else}
        <a class="btn btn-ghost" href={resolve('/')}>Cancel</a>
      {/if}
      {#if step < STEPS.length - 1}
        <button class="btn btn-primary" onclick={next}>Next</button>
      {:else}
        <button class="btn btn-primary" onclick={create} disabled={saving}>{saving ? 'Saving…' : isNew ? 'Create repository' : 'Add repository'}</button>
      {/if}
    </div>
  </div>
</div>
