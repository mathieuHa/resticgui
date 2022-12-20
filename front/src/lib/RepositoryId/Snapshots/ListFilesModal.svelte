
<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { page } from '$app/stores';
  import Modal from '$lib/Modal.svelte';
  import agent from '$utils/api';
  import ListFiles from './ListFiles.svelte';
    import Loader from '$lib/Loader.svelte';

  export let snapshotId: string;

  const dispatch = createEventDispatcher();
  const loadFiles = (async () => agent.get(`/repositories/${$page.params.id}/ls/${snapshotId}`))();
</script>

<Modal title="Files in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
  {#await loadFiles}
    <Loader size={8} />
  {:then files}
    <ListFiles {snapshotId} {files} />
  {/await}
</Modal>