<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import CircleCheck from '@lucide/svelte/icons/circle-check';

  import { api, type Snapshot } from '#lib/api/client.ts';
  import { jobs, notify, notifyError } from '#lib/app.svelte.ts';
  import { isActive } from '#lib/jobs.svelte.ts';
  import { summarize } from '#lib/jobtext.ts';

  import JobProgress from './JobProgress.svelte';
  import Modal from './Modal.svelte';
  import PathPicker from './PathPicker.svelte';

  // Pick where to restore, then follow the restore in the same dialog.
  let {
    open = $bindable(false),
    repositoryId,
    snapshot,
    include = [],
  }: { open?: boolean; repositoryId: number; snapshot: Snapshot; include?: string[] } = $props();

  let where = $state<'folder' | 'original'>('folder');
  let target = $state('');
  let busy = $state(false);
  let jobId = $state<number | null>(null);

  const job = $derived(jobId === null ? undefined : jobs.byId[jobId]);
  const into = $derived(where === 'original' ? '/' : target.trim());
  const what = $derived(include.length ? include.join(', ') : 'the whole snapshot');

  $effect(() => {
    // a new dialog starts from the beginning
    if (open) return;
    jobId = null;
    busy = false;
  });

  async function start() {
    busy = true;
    try {
      const started = await api.restore(repositoryId, snapshot.id, into, include);
      jobs.add([started]);
      jobId = started.id;
      notify(`Restoring into ${into}`, 'info', 3000);
    } catch (err) {
      notifyError(err, 'Could not start the restore');
    } finally {
      busy = false;
    }
  }
</script>

<Modal bind:open title="Restore {what}" wide>
  {#if !job}
    <fieldset class="mb-4 flex flex-col gap-2">
      <legend class="mb-2 text-sm font-medium">Where to</legend>
      <label class="flex cursor-pointer items-start gap-3">
        <input type="radio" class="radio radio-sm radio-primary mt-0.5" value="folder" bind:group={where} />
        <span>
          <span class="font-medium">Another folder</span>
          <span class="text-base-content/70 block text-sm">Files keep their full path under it, such as /restore/home/me/notes.md.</span>
        </span>
      </label>
      <label class="flex cursor-pointer items-start gap-3">
        <input type="radio" class="radio radio-sm radio-primary mt-0.5" value="original" bind:group={where} />
        <span>
          <span class="font-medium">Where they were</span>
          <span class="text-base-content/70 block text-sm">Puts the files back in place. Files with the same name are overwritten.</span>
        </span>
      </label>
    </fieldset>
    {#if where === 'folder'}
      <label class="floating-label mb-3">
        <span>Restore into</span>
        <input class="input w-full font-mono" placeholder="Restore into, such as /restore" bind:value={target} />
      </label>
      <PathPicker onpick={(p) => (target = p)} label="Restore here" />
    {:else}
      <div class="alert alert-warning text-sm" role="note">Existing files at the original paths are replaced by the ones from {snapshot.short_id}.</div>
    {/if}
  {:else if isActive(job.status)}
    <p class="text-base-content/70 mb-3 text-sm">Restoring into <span class="font-mono">{into}</span>. You can close this; the restore keeps going under Runs.</p>
    <JobProgress {job} />
  {:else if job.status === 'succeeded'}
    <div class="animate-rise flex items-center gap-3" role="status">
      <CircleCheck class="text-success size-8 shrink-0" aria-hidden="true" />
      <div>
        <p class="font-medium">Restored</p>
        <p class="text-base-content/70 text-sm">{summarize(job)}</p>
      </div>
    </div>
  {:else}
    <div class="flex items-center gap-3" role="alert">
      <CircleAlert class="text-error size-8 shrink-0" aria-hidden="true" />
      <div>
        <p class="font-medium">The restore {job.status}</p>
        {#if job.error}<p class="text-error text-sm">{job.error}</p>{/if}
      </div>
    </div>
  {/if}
  {#snippet actions()}
    {#if !job}
      <button class="btn" onclick={() => (open = false)}>Cancel</button>
      <button class="btn btn-primary" onclick={start} disabled={busy || !into}>Restore</button>
    {:else}
      <button class="btn" onclick={() => (open = false)}>Close</button>
    {/if}
  {/snippet}
</Modal>
