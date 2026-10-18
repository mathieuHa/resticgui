<script lang="ts">
  import { untrack } from 'svelte';

  import { api, type Plan, type Repository } from '#lib/api/client.ts';
  import { attempt, notify } from '#lib/app.svelte.ts';
  import Modal from '#lib/components/Modal.svelte';
  import PlanEditor from '#lib/components/PlanEditor.svelte';
  import { draftOfPlan, newPlanDraft, planProblems, toPlanInput } from '#lib/plan.ts';

  // One plan of a repository, or a new one when plan is not given.
  let {
    repo,
    plan,
    busy = false,
    onsaved,
    oncancel,
  }: {
    repo: Repository;
    plan?: Plan;
    busy?: boolean;
    /** called with the repository as it is after the change */
    onsaved: (r: Repository) => void;
    oncancel?: () => void;
  } = $props();

  const initial = untrack(() => plan);
  let draft = $state(initial ? draftOfPlan(initial) : newPlanDraft(''));
  let errors = $state<string[]>([]);
  let confirmDelete = $state(false);
  const others = $derived(repo.plans.filter((p) => p.id !== plan?.id).map((p) => p.name));

  async function save() {
    errors = planProblems(draft, others);
    if (errors.length) return;
    const input = toPlanInput($state.snapshot(draft));
    await attempt(async () => {
      if (plan) await api.updatePlan(repo.id, plan.id, input);
      else await api.createPlan(repo.id, input);
      onsaved(await api.repository(repo.id));
      notify(plan ? `Plan ${input.name} saved` : `Plan ${input.name} added`);
    }, 'Could not save the plan');
  }

  async function remove() {
    if (!plan) return;
    const name = plan.name;
    if (await attempt(() => api.deletePlan(repo.id, plan.id), 'Could not delete the plan')) {
      confirmDelete = false;
      onsaved(await api.repository(repo.id));
      notify(`Plan ${name} deleted. Its snapshots are kept.`);
    }
  }
</script>

<section class="card bg-base-100 border-base-300 border shadow-sm" data-testid="plan-settings" aria-label={plan ? `Plan ${plan.name}` : 'New plan'}>
  <div class="card-body">
    <h3 class="card-title text-base">{plan ? plan.name : 'New plan'}</h3>
    {#if errors.length}
      <ul class="text-error text-sm">
        {#each errors as e (e)}<li>{e}</li>{/each}
      </ul>
    {/if}
    <PlanEditor bind:draft repositoryId={repo.id} planId={plan?.id} />
    <div class="card-actions justify-end">
      {#if plan}
        <button class="btn btn-ghost btn-sm text-error" onclick={() => (confirmDelete = true)} disabled={busy} title={busy ? 'Wait for the running job to finish' : ''}>
          Delete plan
        </button>
      {:else}
        <button class="btn btn-ghost btn-sm" onclick={() => oncancel?.()}>Cancel</button>
      {/if}
      <button class="btn btn-primary btn-sm" onclick={save}>{plan ? 'Save plan' : 'Add plan'}</button>
    </div>
  </div>
</section>

<Modal bind:open={confirmDelete} title="Delete the plan {plan?.name}?">
  <p>Its snapshots stay in the repository, and are no longer removed by a retention policy.</p>
  {#snippet actions()}
    <button class="btn" onclick={() => (confirmDelete = false)}>Cancel</button>
    <button class="btn btn-error" onclick={remove}>Delete</button>
  {/snippet}
</Modal>
