<script>
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';

  import agent from '$utils/api';
  import Eye from '$lib/Icon/Eye.svelte';
  import Logs from '$lib/Icon/Logs.svelte';
  import Pen from '$lib/Icon/Pen.svelte';
  import Trash from '$lib/Icon/Trash.svelte';
  import Stats from '$lib/Icon/Stats.svelte';
  import RepositoryModal from '$lib/RepositoryModal.svelte';
  import Modal from '$lib/Modal.svelte';
  import Snapshot from '$lib/Icon/Snapshot.svelte';

  let repos = [];
  let step = 0;
  let initialValues = null;
  let removeId = null;

  const onRefresh = async () => {
    try {
      repos = await agent.get('/repositories');
      initialValues = repos.find((repo) => repo.id === initialValues.id);
    } catch (err) {
      // Pass
    }
  }
  
  const onRemove = async (repoId) => {
    try {
      await agent.delete(`/repositories/${repoId}`);
      await onRefresh();
      removeId = null;
    } catch (err) {
      console.log(err)
    }
  }

  $:if (initialValues === null) step = 0;
  $:values = Object.assign({}, initialValues) || step;

  onMount(onRefresh);
</script>

{#if removeId}
  <Modal title="Advertissment" on:close={() => { removeId = null; }}>
    <div>
      You really want to delete repository #{removeId}
      <div class="flex gap-2 justify-end">
        <button class="bg-blue-500 hover:bg-blue-700 text-white py-1 px-4 rounded transition" on:click={() => removeId = null}>No</button>
        <button class="bg-red-400 hover:bg-red-600 text-white py-1 px-4 rounded transition" on:click={() => onRemove(removeId)}>Yes</button>
      </div>
    </div>
  </Modal>
{:else if initialValues}
  <RepositoryModal
    repository={initialValues}
    on:close={() => { initialValues = null; }}
    on:refresh={onRefresh}
  />
{/if}

<div class="w-full grid gap-3 mb-6">
  {#each repos as repo, i}
    <div class="p-3 flex flex-col bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <div class="flex flex-wrap gap-2">
        <p class="text-lg font-bold leading-7 whitespace-nowrap {repo.active ? '' :'text-red-500'}">#{repo.id} - {repo.name}</p>
        <p class="flex-1 text-sm font-light leading-7 text-right whitespace-nowrap ">{dayjs(repo.createdAt).format('DD/MM/YYYY HH:mm ')}</p>
      </div>

      <div class="flex flex-wrap gap-2 mt-3 pt-2 min-h-8 mb-1">
        <a href="/{repo.id}">
          <button
            class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-yellow-500 hover:bg-yellow-600"
            title="See"
          >
            <Eye />
          </button>
        </a>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          title="Update"
          on:click|preventDefault|stopPropagation={() => { initialValues = Object.assign({}, repo); }}
        >
          <Pen />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-red-500 hover:bg-red-600"
          title="Remove"
          on:click|preventDefault|stopPropagation={() => { removeId = repo.id}}
        >
          <Trash />
        </button>
        <a href="/{repo.id}?tab=logs">
          <button
            class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-orange-500 hover:bg-orange-600"
            title="Logs"
          >
            <Logs />
          </button>
        </a>
        <a href="/{repo.id}?tab=snapshots">
          <button
            class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-violet-500 hover:bg-violet-600"
            title="Snapshots"
          >
            <Snapshot />
          </button>
        </a>
        <a href="/{repo.id}?tab=stats">
          <button
            class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-green-500 hover:bg-green-600"
            title="Stats"
          >
            <Stats />
          </button>
        </a>
      </div>
    </div>
  {/each}
  <button
    on:click={() => { initialValues = {}; }}
    class="p-6 bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center flex justify-center cursor-pointer"
  >
    Add a repository
  </button>
</div>
