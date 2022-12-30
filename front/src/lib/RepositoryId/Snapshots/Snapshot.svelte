<script lang="ts">
  import dayjs from 'dayjs';
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import Eye from '$lib/Icon/Eye.svelte';
  import Modal from '$lib/Modal.svelte';
  import Trash from '$lib/Icon/Trash.svelte';
  import Restore from '$lib/Icon/Restore.svelte';
  import RestoreModal from './RestoreModal.svelte';
  import SeeModal from './SeeModal.svelte';
  import FilesModal from './FilesModal.svelte';
  import List from '$lib/Icon/List.svelte';

  interface Snapshot {
    short_id: string
    username: string
    time: string
    Paths: string[]
  }
  export let snapshots: Snapshot[] = [];

  let snapshot: Snapshot | null = null;
  let modalOpened: string | null = null;
  const dispatch = createEventDispatcher();

  const deleteSnapshot = async (snapshotId: string) => {
    const id = $page.params.id;
    await agent.delete(`/repositories/${id}/snapshots/${snapshotId}`);
    dispatch('delete');
  };
  const setModalOpened = (rowSnapshot: any, toOpen: string) => {
    snapshot = rowSnapshot;
    modalOpened = toOpen;
  };

  $: snapshotsList = snapshots.sort((a: Snapshot, b: Snapshot) => (dayjs(b.time).valueOf() - dayjs(a.time).valueOf()));
</script>
  
{#if snapshot}
  {#if modalOpened === 'see'}
    <SeeModal snapshotId={snapshot.short_id} paths={snapshot.Paths} on:close={() => (modalOpened = null)}/>
  {:else if modalOpened === 'delete'}
    <Modal title="Warning" on:close={() => (modalOpened = null)}>
      <div>
        Do you really want to delete this snapshot #{snapshot.short_id}
        <div class="flex gap-2 justify-end">
          <button
            class="bg-blue-500 hover:bg-blue-700 text-white py-1 px-4 rounded transition"
            on:click={() => (modalOpened = null)}
          >
            No
          </button>
          <button
            class="bg-red-400 hover:bg-red-600 text-white py-1 px-4 rounded transition"
            on:click={() => snapshot && deleteSnapshot(snapshot.short_id)}
          >
            Yes
          </button>
        </div>
      </div>
    </Modal>
  {:else if modalOpened === 'restore'}
    <RestoreModal snapshotId={snapshot.short_id} on:close={() => (modalOpened = null)} />
  {:else if modalOpened === 'files'}
    <FilesModal snapshotId={snapshot.short_id}  on:close={() => (modalOpened = null)} />
  {/if}
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each snapshotsList as snapshot}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p class="truncate text-lg font-bold">#{snapshot.short_id}</p>
      <p class="truncate text-sm">{snapshot.username}</p>
      <p class="truncate text-sm font-light my-2">{dayjs(snapshot.time).format('DD/MM/YYYY HH:mm ')}</p>
      <div class="flex flex-wrap gap-2 mt-auto pt-2 min-h-8 mb-1">
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-yellow-500 hover:bg-yellow-600"
          title="See paths"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'see')}
        >
          <Eye />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-red-500 hover:bg-red-600"
          title="Delete snapshot"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'delete')}
        >
          <Trash />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-cyan-500 hover:bg-cyan-600"
          title="Restore snapshot"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'restore')}
        >
          <Restore />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-green-500 hover:bg-green-600"
          title="Snapshot files"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'files')}
        >
          <List />
        </button>
      </div>
    </div>
  {/each}
</div>