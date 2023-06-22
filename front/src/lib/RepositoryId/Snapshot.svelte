<script lang="ts">
  import dayjs from 'dayjs';
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import Eye from '$lib/Icon/Eye.svelte';
  import Trash from '$lib/Icon/Trash.svelte';
  import Restore from '$lib/Icon/Restore.svelte';
  import RestoreModal from './Snapshots/RestoreModal.svelte';
  import SeeModal from './Snapshots/SeeModal.svelte';
  import FilesModal from './Snapshots/FilesModal.svelte';
  import List from '$lib/Icon/List.svelte';
  import { toasts } from '$utils/stores';
    import DeleteModal from './Snapshots/DeleteModal.svelte';

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

  const setModalOpened = (rowSnapshot: any, toOpen: string) => {
    snapshot = rowSnapshot;
    modalOpened = toOpen;
  };

  $: snapshotsList = snapshots.sort((a: Snapshot, b: Snapshot) => (dayjs(b.time).valueOf() - dayjs(a.time).valueOf()));
</script>
  
{#if snapshot}
  {#if modalOpened === 'see'}
    <SeeModal
      snapshotId={snapshot.short_id}
      paths={snapshot.Paths}
      on:close={() => (modalOpened = null)}
    />
  {:else if modalOpened === 'delete'}
    <DeleteModal
      snapshotId={snapshot.short_id}
      on:close={() => (modalOpened = null)}
      on:delete={() => dispatch('delete')}
    />
  {:else if modalOpened === 'restore'}
    <RestoreModal
      snapshotId={snapshot.short_id}
      on:close={() => (modalOpened = null)}
    />
  {:else if modalOpened === 'files'}
    <FilesModal
      snapshotId={snapshot.short_id}
      on:close={() => (modalOpened = null)}
    />
  {/if}
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each snapshotsList as snapshot}
    <div class="p-3 flex flex-col max-w-sm bg-base-100 rounded-xl shadow-md">
      <p class="truncate text-lg font-bold">#{snapshot.short_id}</p>
      <p class="truncate text-sm">{snapshot.username}</p>
      <p class="truncate text-sm font-light my-2">{dayjs(snapshot.time).format('DD/MM/YYYY HH:mm')}</p>
      <div class="flex flex-wrap gap-2 mt-auto pt-2 min-h-8 mb-1">
        <button
          class="btn btn-square btn-sm btn-warning"
          title="See paths"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'see')}
        >
          <Eye />
        </button>
        <button
          class="btn btn-square btn-sm btn btn-error"
          title="Delete snapshot"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'delete')}
        >
          <Trash />
        </button>
        <button
          class="btn btn-square btn-sm btn btn-info"
          title="Restore snapshot"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'restore')}
        >
          <Restore />
        </button>
        <button
          class="btn btn-square btn-sm btn btn-success"
          title="Snapshot files"
          on:click|preventDefault|stopPropagation={() => setModalOpened(snapshot, 'files')}
        >
          <List />
        </button>
      </div>
    </div>
  {/each}
</div>