<script lang="ts">
  import { fly } from 'svelte/transition';

  import { ms } from '#lib/motion.ts';
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { api, type Backend, type Probe } from '#lib/api/client.ts';
  import { jobs, notify, notifyError } from '#lib/app.svelte.ts';
  import PlanEditor from '#lib/components/PlanEditor.svelte';
  import RepositoryForm from '#lib/components/RepositoryForm.svelte';
  import ScheduleEditor from '#lib/components/ScheduleEditor.svelte';
  import { newPlanDraft, planProblems, toPlanInput } from '#lib/plan.ts';
  import { BACKEND_LABELS, newDraft, problems, toInput } from '#lib/repository.ts';
  import { describeRetention } from '#lib/retention.ts';
  import { describe, parse } from '#lib/schedule.ts';

  const STEPS = ['Storage', 'What to back up', 'When', 'Review'];
  let step = $state(0);
  let backends = $state<Backend[]>([]);
  let draft = $state(newDraft());
  let isNew = $state(true);
  let plan = $state(newPlanDraft());
  let checkCron: string | null = $state(null);
  let saving = $state(false);
  let shown = $state<string[]>([]);
  let probing = $state(false);
  let found = $state<Probe | null>(null);
  let prefilled = $state(false);

  $effect(() => {
    api.backends().then(
      (b) => (backends = b),
      (err) => notifyError(err, 'Could not load the storage types'),
    );
  });

  const backend = $derived(backends.find((b) => b.name === draft.backend));
  const stepProblems = $derived.by(() => {
    if (step === 0) return problems(draft, backend, true);
    if (step === 1 || step === 2) return planProblems(plan);
    return [];
  });

  async function next() {
    shown = stepProblems;
    if (shown.length) return;
    if (step === 0 && !(await look())) return;
    step++;
  }

  // what is at the location: an existing repository is connected, never initialized again
  async function look(): Promise<boolean> {
    probing = true;
    try {
      const probe = await api.probe(toInput(draft, backend));
      if (isNew && probe.exists) {
        shown = ['A restic repository already exists at this location. Choose "Use an existing one" to connect it.'];
        return false;
      }
      if (!isNew && !probe.exists) {
        shown = ['No restic repository was found at this location. Check it, or choose "Create a new repository".'];
        return false;
      }
      found = isNew ? null : probe;
      const latest = probe.latest;
      if (latest && !plan.paths.include.length) {
        plan.paths = { include: [...latest.paths], exclude: [] };
        prefilled = true;
      }
      if (latest && !draft.host.trim()) draft.host = latest.hostname;
      return true;
    } catch (err) {
      shown = [err instanceof Error ? err.message : String(err)];
      return false;
    } finally {
      probing = false;
    }
  }

  async function create() {
    saving = true;
    let id: number | null = null;
    try {
      const repo = await api.createRepository(toInput(draft, backend));
      id = repo.id;
      await api.createPlan(id, toPlanInput($state.snapshot(plan)));
      if (checkCron) await api.setCheckSchedule(id, checkCron);
      if (isNew) jobs.add([await api.startJob(id, 'init')]);
      // listing the snapshots marks a connected repository ready
      else await api.snapshots(id, true);
      notify(isNew ? `${repo.name} is being created` : `${repo.name} connected`);
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

<div class="card bg-base-100 border-base-300 animate-rise mx-auto max-w-3xl border shadow-sm">
  <div class="card-body gap-6">
    <h1 class="card-title">New repository</h1>
    <ul class="steps w-full">
      {#each STEPS as name, i (name)}
        <li class="step {i <= step ? 'step-primary' : ''}">{name}</li>
      {/each}
    </ul>

    {#key step}
    <div class="flex flex-col gap-6" in:fly={{ x: 16, duration: ms(200) }}>
    {#if step === 0}
      <div class="join">
        <button class="btn join-item btn-sm {isNew ? 'btn-active' : ''}" onclick={() => { isNew = true; shown = []; }}>Create a new repository</button>
        <button class="btn join-item btn-sm {!isNew ? 'btn-active' : ''}" onclick={() => { isNew = false; shown = []; }}>Use an existing one</button>
      </div>
      <RepositoryForm bind:draft {backends} />
    {:else if step === 1}
      {#if found}
        <div class="alert alert-info text-sm" data-testid="found">
          {#if found.latest}
            Found {found.snapshots} {found.snapshots === 1 ? 'snapshot' : 'snapshots'}.
            {#if prefilled}The folders of the latest one, from {found.latest.hostname}, are filled in.{/if}
          {:else}
            Found the repository. It has no snapshot yet.
          {/if}
        </div>
      {/if}
      <p class="text-base-content/70 text-sm">
        A plan says what to back up and when. You can add more plans later, such as photos every day and documents every hour, all in this repository.
      </p>
      <PlanEditor bind:draft={plan} show="what" />
    {:else if step === 2}
      <PlanEditor bind:draft={plan} show="when" />
      <div class="divider my-0"></div>
      <ScheduleEditor label="Check the repository's integrity" bind:value={checkCron} />
    {:else}
      <dl class="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
        <dt class="text-base-content/70">Name</dt><dd>{draft.name}</dd>
        <dt class="text-base-content/70">Storage</dt><dd>{BACKEND_LABELS[draft.backend]} · <span class="font-mono">{draft.location}</span></dd>
        <dt class="text-base-content/70">Repository</dt><dd>{isNew ? 'Created now with restic init' : `Already exists, with ${found?.snapshots ?? 0} snapshots`}</dd>
        <dt class="text-base-content/70">Plan</dt><dd>{plan.name}</dd>
        <dt class="text-base-content/70">Backs up</dt><dd class="font-mono">{plan.paths.include.join(', ')}</dd>
        {#if plan.paths.exclude.length}<dt class="text-base-content/70">Excludes</dt><dd class="font-mono">{plan.paths.exclude.join(', ')}</dd>{/if}
        <dt class="text-base-content/70">Backups</dt><dd>{plan.backupCron ? describe(parse(plan.backupCron)) : 'Only when started by hand'}</dd>
        <dt class="text-base-content/70">Retention</dt>
        <dd>{plan.retentionOn ? describeRetention(plan.retention) : 'Keep everything'}{plan.retentionOn && plan.forgetCron ? `, ${describe(parse(plan.forgetCron)).toLowerCase()}` : ''}</dd>
        <dt class="text-base-content/70">Check</dt><dd>{checkCron ? describe(parse(checkCron)) : 'Only when started by hand'}</dd>
      </dl>
    {/if}
    </div>
    {/key}

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
        <button class="btn btn-primary" onclick={next} disabled={probing}>{probing ? 'Looking at the location…' : 'Next'}</button>
      {:else}
        <button class="btn btn-primary" onclick={create} disabled={saving}>{saving ? 'Saving…' : isNew ? 'Create repository' : 'Add repository'}</button>
      {/if}
    </div>
  </div>
</div>
