
<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { page } from '$app/stores';
  import Modal from '$lib/Modal.svelte';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';
  import ListFiles from './ListFiles.svelte';

  export let snapshotId: string;
  let files: any[] = []

  const dispatch = createEventDispatcher();
  onMount(
    async () => {
      try {
        files = await agent.get(`/repositories/${$page.params.id}/ls/${snapshotId}`);
        toasts.addToast(`List files of ${snapshotId} received`);
      // dispatch('close');
      } catch (err) {
        toasts.addToast(err as string, 'failure');
      }
    }
  )
  
</script>

<Modal title="Files in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
  <div>
    <ListFiles {snapshotId} {files} />
  </div>
</Modal>