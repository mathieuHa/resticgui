<script lang="ts">
  import dayjs from 'dayjs';
  import { createEventDispatcher } from 'svelte';
  import {page} from '$app/stores';
  import agent from '$utils/api';
  import Eye from '../Icon/Eye.svelte';
  import Modal from '../Modal.svelte';
  import Trash from '../Icon/Trash.svelte';
  import Restore from '../Icon/Restore.svelte';
  import List from '../Icon/List.svelte';
  import Search from '../Icon/Search.svelte';
  import RestoreModal from './RestoreModal.svelte';
  import ListFilesModal from './ListFilesModal.svelte';
  import SearchFilesModal from './SearchFilesModal.svelte';
  export let snapshots: any = [];

  let visible: any = null;
  let removeId: string | null = null;
  let restoreId: string | null = null;
  let snapshotId: string | null = null;
  let searchId: string | null = null;
  const dispatch = createEventDispatcher();

  const deleteSnapshot = async (snapshotId: string) => {
    const id = $page.params.id;
    await agent.delete(`/repositories/${id}/snapshots/${snapshotId}`);
    dispatch('delete');
  };
  
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
{:else if removeId !== null}
  <Modal title="Warning" on:close={() => (removeId = null)}>
    <div>
      Do you really want to delete this snapshot #{removeId}
      <div class="flex gap-2 justify-end">
        <button
          class="bg-blue-500 hover:bg-blue-700 text-white py-1 px-4 rounded transition"
          on:click={() => (removeId = null)}
        >
          No
        </button>
        <button
          class="bg-red-400 hover:bg-red-600 text-white py-1 px-4 rounded transition"
          on:click={() => removeId && deleteSnapshot(removeId)}
        >
          Yes
        </button>
      </div>
    </div>
  </Modal>
{:else if restoreId !== null}
  <RestoreModal {restoreId} on:close={() => (restoreId = null)} />
{:else if snapshotId !== null}
  <ListFilesModal {snapshotId} on:close={() => (snapshotId = null)} />
{:else if searchId !== null}
  <SearchFilesModal {searchId} on:close={() => (searchId = null)} />
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
          title="See paths"
          on:click|preventDefault|stopPropagation={() => (visible = snapshot)}
        >
          <Eye />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-red-500 hover:bg-red-600"
          title="delete snapshot"
          on:click|preventDefault|stopPropagation={() => (removeId = snapshot.short_id)}
        >
          <Trash />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-cyan-500 hover:bg-cyan-600"
          title="restore snapshot"
          on:click|preventDefault|stopPropagation={() => (restoreId = snapshot.short_id)}
        >
          <Restore />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-orange-500 hover:bg-orange-600"
          title="list snapshot files"
          on:click|preventDefault|stopPropagation={() => (snapshotId = snapshot.short_id)}
        >
          <List />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-green-500 hover:bg-green-600"
          title="search snapshot files"
          on:click|preventDefault|stopPropagation={() => (searchId = snapshot.short_id)}
        >
          <Search />
        </button>
      </div>
    </div>
  {/each}
</div>