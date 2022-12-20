<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import { toasts } from '$utils/stores';
  import Modal from '$lib/Modal.svelte';
  import agent from '$utils/api';
  import Input from '$lib/Input.svelte';
  import Search from '$lib/Icon/Search.svelte';
  import ListFiles from './ListFiles.svelte';
    import Loader from '$lib/Loader.svelte';

  export let snapshotId: string;
  let files: any[] = []
  let pattern: string | null = null;

  const dispatch = createEventDispatcher();
  const search = async () => agent.get(`/repositories/${$page.params.id}/search/${snapshotId}?pattern=${pattern}`);
  let searchPromise: Promise<any> | null;
</script>

<Modal title="Files found in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
  <div class="flex gap-2">
    <div class="flex-1">
      <Input
        secondary
        id="search"
        label="search file"
        type="text"
        bind:value={pattern}
      />
    </div>
    <button
      class="mt-auto mb-2 rounded h-8 w-8 text-gray-900 dark:text-gray-100 p-2 transition bg-blue-500 hover:bg-blue-600"
      on:click={() => (searchPromise = search())}
    >
      <Search />
    </button>
  </div>
  
  {#await searchPromise}
    <Loader size={8} />
  {:then files}
    <ListFiles {snapshotId} {files} {pattern} />
  {/await}
</Modal>