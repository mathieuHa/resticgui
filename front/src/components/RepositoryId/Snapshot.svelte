<script>
  import dayjs from 'dayjs';

  import Eye from '../Icon/Eye.svelte';
  import Modal from '../Modal.svelte';

  export let snapshots = [];

  let visible = null;

</script>

{#if visible}
  <Modal
    title="Paths of {visible.short_id} at {dayjs(visible.time).format('DD/MM/YYYY HH:mm ')}"
    on:close={() => { visible = null; }}
  >
    {#each visible.Paths as path}
      <p>{path}</p>
    {/each}
  </Modal>
{/if}
<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each snapshots as snapshot}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p class="truncate text-lg font-bold">#{snapshot.short_id}</p>
      <p class="truncate text-sm">{snapshot.username}</p>
      <p class="truncate text-sm font-light my-2">{dayjs(snapshot.time).format('DD/MM/YYYY HH:mm ')}</p>
      <div class="flex flex-wrap gap-2 mt-auto pt-2 min-h-8 mb-1">
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-yellow-500 hover:bg-yellow-600"
          title="Restore"
          on:click={() =>  { visible = snapshot; }}
        >
          <Eye />
        </button>
      </div>
    </div>
  {/each}
</div>