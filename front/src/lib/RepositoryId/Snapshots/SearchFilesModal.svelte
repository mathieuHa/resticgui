<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import { toasts } from '$utils/stores';
  import Modal from '$lib/Modal.svelte';
  import agent from '$utils/api';
  import Input from '$lib/Input.svelte';
  import Search from '$lib/Icon/Search.svelte';
  import ListFiles from './ListFiles.svelte';

  export let snapshotId: string | null = null;
  let files: any[] = []
  let pattern: string | null = null;

  const dispatch = createEventDispatcher();
  const search = async () => {
      try {
        files = await agent.get(`/repositories/${$page.params.id}/search/${snapshotId}?pattern=${pattern}`);
        toasts.addToast(`Search pattern of filenames of ${snapshotId} received`);
      // dispatch('close');
      } catch (err) {
        toasts.addToast(err as string, 'failure');
      }
    }
  
</script>

<Modal title="Files found in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
  <div>
    <div class="grid grid-flow-col">
      <div class="col">
        <Input
          id="search"
          label="search file"
          type="text"
          bind:value={pattern}
        />
      </div>
      <div class="col">
        <button
          class="bg-blue-500 hover:bg-blue-700 text-white py-2 px-2 rounded transition text-right align-middle"
          on:click={() => search()}>
          <Search />
        </button>
      </div>
    </div>
    
    <ListFiles {snapshotId} {files} {pattern} />
  </div>
</Modal>