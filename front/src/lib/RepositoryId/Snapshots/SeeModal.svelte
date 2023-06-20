
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import Number from '$lib/Number.svelte';
  import Loader from '$lib/Loader.svelte';
    import Modal from '$lib/Modal.svelte';

  export let snapshotId: string | null = null;
  export let paths: string[] = [];

  const dispatch = createEventDispatcher();
  const loadStats = (async () => agent.get(`/repositories/${$page.params.id}/stats?snapshotID=${snapshotId}`))();
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<Modal on:close={() => dispatch('close')}>
  <div class="modal-box" on:click|preventDefault|stopPropagation={() => {}}>
    {#await loadStats}
      <Loader size={8} />
    {:then stats}
      <h3 class="font-bold text-lg mb-4">Stats and path saved in snapshot #{snapshotId}</h3>
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
        <ul>
          {#each paths as path}
            <li>{path}</li>
          {/each}
        </ul>
      </div>
    {:catch err}
      <pre>{err}</pre>
    {/await}
  </div>
</Modal>