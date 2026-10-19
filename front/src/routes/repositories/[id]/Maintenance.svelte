<script lang="ts">
  import { api, type Job, type Repository } from '#lib/api/client.ts';
  import { attempt, jobs, notify } from '#lib/app.svelte.ts';
  import Modal from '#lib/components/Modal.svelte';
  import { KIND_LABELS, SUBSETS } from '#lib/jobtext.ts';

  let { repo }: { repo: Repository } = $props();

  type Tool = 'repair_index' | 'repair_snapshots' | 'migrate';
  // what each tool's confirmation says
  const CONFIRM: Record<Tool, { title: string; text: string; action: string }> = {
    repair_index: {
      title: 'Rebuild the index?',
      text: 'restic reads the list of every pack file in the repository and writes a new index from it. Nothing is lost, but it can take long on a large or remote repository.',
      action: 'Rebuild',
    },
    repair_snapshots: {
      title: 'Repair the snapshots?',
      text: 'Snapshots that refer to data the repository lost are saved again without it, and the damaged ones are forgotten. The files that cannot be read are gone from these snapshots for good. Rebuild the index first.',
      action: 'Repair',
    },
    migrate: {
      title: 'Upgrade the repository format?',
      text: 'restic checks the repository, then upgrades it to format version 2, which compresses new data. restic older than 0.14 can no longer open it, and there is no way back. A repository already at version 2 is left as it is.',
      action: 'Upgrade',
    },
  };

  let subset = $state('');
  let confirming = $state<Tool | null>(null);
  let password = $state('');
  let repeat = $state('');
  let changing = $state(false);
  const mismatch = $derived(repeat !== '' && password !== repeat);

  function started(job: Job, what: string) {
    jobs.add([job]);
    notify(`${what} started`, 'info', 3000);
  }

  async function check() {
    await attempt(async () => started(await api.startCheck(repo.id, subset), KIND_LABELS.check), 'Could not start the check');
  }

  async function unlock() {
    await attempt(async () => started(await api.startJob(repo.id, 'unlock'), KIND_LABELS.unlock), 'Could not remove the locks');
  }

  async function run(tool: Tool) {
    confirming = null;
    await attempt(async () => started(await api.startJob(repo.id, tool), KIND_LABELS[tool]), `Could not start: ${KIND_LABELS[tool]}`);
  }

  async function changePassword() {
    changing = false;
    const ok = await attempt(async () => started(await api.changePassword(repo.id, password), KIND_LABELS.change_password), 'Could not change the password');
    if (ok) password = repeat = '';
  }
</script>

<section class="card bg-base-100 border-base-300 border shadow-sm" data-testid="maintenance">
  <div class="card-body gap-0">
    <h2 class="card-title text-base">Maintenance</h2>
    <p class="text-base-content/70 mb-2 text-sm">Each tool runs as a job, after the jobs already waiting for this repository. Its result is in Runs.</p>

    <div class="border-base-300 flex flex-col gap-2 border-t py-3 sm:flex-row sm:items-center">
      <div class="flex-1">
        <h3 class="font-medium">Check now</h3>
        <p class="text-base-content/70 text-sm">Verifies the structure of the repository and, if you choose, reads some of the data to find damaged files.</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <select class="select select-sm w-auto" bind:value={subset} aria-label="Data the check reads">
          {#each SUBSETS as [value, label] (value)}<option {value}>{label}</option>{/each}
        </select>
        <button class="btn btn-sm" onclick={check}>Check now</button>
      </div>
    </div>

    <div class="border-base-300 flex flex-col gap-2 border-t py-3 sm:flex-row sm:items-center">
      <div class="flex-1">
        <h3 class="font-medium">Remove stale locks</h3>
        <p class="text-base-content/70 text-sm">A restic that was stopped or crashed can leave a lock that blocks the next runs. Locks of commands still running are kept.</p>
      </div>
      <button class="btn btn-sm" onclick={unlock}>Remove stale locks</button>
    </div>

    <div class="border-base-300 flex flex-col gap-2 border-t py-3 sm:flex-row sm:items-center">
      <div class="flex-1">
        <h3 class="font-medium">Rebuild the index</h3>
        <p class="text-base-content/70 text-sm">When a check reports index errors, or files were lost on the storage.</p>
      </div>
      <button class="btn btn-sm" onclick={() => (confirming = 'repair_index')}>Rebuild the index</button>
    </div>

    <div class="border-base-300 flex flex-col gap-2 border-t py-3 sm:flex-row sm:items-center">
      <div class="flex-1">
        <h3 class="font-medium">Repair snapshots</h3>
        <p class="text-base-content/70 text-sm">After rebuilding the index, saves the snapshots that refer to lost data again without it, so that the others can be restored and checked.</p>
      </div>
      <button class="btn btn-sm btn-warning" onclick={() => (confirming = 'repair_snapshots')}>Repair snapshots</button>
    </div>

    <div class="border-base-300 flex flex-col gap-2 border-t py-3 sm:flex-row sm:items-center">
      <div class="flex-1">
        <h3 class="font-medium">Upgrade the format</h3>
        <p class="text-base-content/70 text-sm">Repositories made by restic before 0.14 use format version 1, without compression. Version 2 compresses new data.</p>
      </div>
      <button class="btn btn-sm btn-warning" onclick={() => (confirming = 'migrate')}>Upgrade the format</button>
    </div>

    <form
      class="border-base-300 flex flex-col gap-2 border-t pt-3"
      onsubmit={(e) => {
        e.preventDefault();
        if (password && !mismatch) changing = true;
      }}
    >
      <h3 class="font-medium">Change the password</h3>
      <p class="text-base-content/70 text-sm">
        The new password opens the repository in place of the current one. Keep it somewhere safe: without it, the backups cannot be read.
      </p>
      <div class="flex flex-col gap-2 sm:flex-row sm:items-end">
        <label class="flex flex-1 flex-col gap-1">
          <span class="text-sm">New password</span>
          <input class="input input-sm w-full" type="password" autocomplete="new-password" bind:value={password} />
        </label>
        <label class="flex flex-1 flex-col gap-1">
          <span class="text-sm">Repeat it</span>
          <input class="input input-sm w-full {mismatch ? 'input-error' : ''}" type="password" autocomplete="new-password" bind:value={repeat} aria-invalid={mismatch} />
        </label>
        <button class="btn btn-sm btn-warning" type="submit" disabled={!password || password !== repeat}>Change password</button>
      </div>
      {#if mismatch}<p class="text-error text-sm">The two passwords differ.</p>{/if}
    </form>
  </div>
</section>

<Modal open={!!confirming} title={confirming ? CONFIRM[confirming].title : ''} onclose={() => (confirming = null)}>
  <p>{confirming ? CONFIRM[confirming].text : ''}</p>
  {#snippet actions()}
    <button class="btn" onclick={() => (confirming = null)}>Cancel</button>
    <button class="btn {confirming === 'repair_index' ? 'btn-primary' : 'btn-warning'}" onclick={() => confirming && run(confirming)}>
      {confirming ? CONFIRM[confirming].action : ''}
    </button>
  {/snippet}
</Modal>

<Modal bind:open={changing} title="Change the password of {repo.name}?">
  <p class="mb-2">restic adds a key for the new password, ResticGUI saves it, then the key of the current password is removed.</p>
  <p>Keys made for other computers stay. Anything else that opens this repository with the current password, such as a script, stops working.</p>
  {#snippet actions()}
    <button class="btn" onclick={() => (changing = false)}>Cancel</button>
    <button class="btn btn-warning" onclick={changePassword}>Change password</button>
  {/snippet}
</Modal>
