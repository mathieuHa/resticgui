
<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { page } from '$app/stores';
    import { toasts } from '$utils/stores';
    import agent from '$utils/api';
    import Modal from '$lib/Modal.svelte';
    import Number from '$lib/Number.svelte';
  
    export let snapshotId: string | null = null;
    export let paths: string[] = [];
    let stats: any = {
      total_file_count: 0,
      total_size: 0,
    }
  
    const dispatch = createEventDispatcher();
    onMount(
      async () => {
        try {
          stats = await agent.get(`/repositories/${$page.params.id}/stats?snapshotID=${snapshotId}`);
          toasts.addToast(`Stats of ${snapshotId} received`);
        } catch (err) {
          toasts.addToast(err as string, 'failure');
        }
      }
    )
    
  </script>
  
  <Modal title="Stats and path saved in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
    <div>
      <p>
        Size:
        <span class="ml-3"><Number number={stats.total_size || 0} /></span>
      </p>
      <p>
        Total file number:
        <span class="ml-3">{stats.total_file_count}</span>
      </p>
      <br />
      Path saved:
      {#each paths as path}
        <p>{path}</p>
      {/each}
    </div>
  </Modal>