<script lang="ts">
  import { api, type Snapshot } from '$lib/api/client';
  import { jobs, notify, notifyError } from '$lib/app.svelte';

  import Modal from './Modal.svelte';
  import PathPicker from './PathPicker.svelte';

  let {
    open = $bindable(false),
    repositoryId,
    snapshot,
    include = [],
  }: { open?: boolean; repositoryId: number; snapshot: Snapshot; include?: string[] } = $props();

  let target = $state('');
  let busy = $state(false);

  async function start() {
    busy = true;
    try {
      const job = await api.restore(repositoryId, snapshot.id, target.trim(), include);
      jobs.add([job]);
      notify(`Restoring into ${target.trim()}. Follow it under Runs.`, 'info');
      open = false;
    } catch (err) {
      notifyError(err, 'Could not start the restore');
    } finally {
      busy = false;
    }
  }
</script>

<Modal bind:open title="Restore {include.length ? include.join(', ') : 'the whole snapshot'}" wide>
  <p class="mb-3 text-sm opacity-70">
    Files are restored under the target folder with their full path, and existing files with the same name are overwritten.
  </p>
  <label class="floating-label mb-3">
    <span>Restore into</span>
    <input class="input w-full font-mono" placeholder="Restore into, such as /restore" bind:value={target} />
  </label>
  <PathPicker onpick={(p) => (target = p)} label="Restore here" />
  {#snippet actions()}
    <button class="btn" onclick={() => (open = false)}>Cancel</button>
    <button class="btn btn-primary" onclick={start} disabled={busy || !target.trim()}>Restore</button>
  {/snippet}
</Modal>
