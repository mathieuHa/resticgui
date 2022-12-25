
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import Modal from '$lib/Modal.svelte';
  import Number from '$lib/Number.svelte';
  import Loader from '$lib/Loader.svelte';

  export let snapshotId: string | null = null;
  export let paths: string[] = [];

  const dispatch = createEventDispatcher();
  const loadStats = (async () => agent.get(`/repositories/${$page.params.id}/stats?snapshotID=${snapshotId}`))();
</script>
  
<Modal title="Stats and path saved in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
  {#await loadStats}
    <Loader size={8} />
  {:then stats}
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
  {:catch err}
    <pre>{err}</pre>
  {/await}
</Modal>