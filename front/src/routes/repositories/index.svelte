<script>
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';

  import agent from '../../utils/api';
  import Eye from '../../components/Icon/Eye.svelte';
  import Logs from '../../components/Icon/Logs.svelte';
  import Pen from '../../components/Icon/Pen.svelte';
  import Trash from '../../components/Icon/Trash.svelte';
  import LogModal from '../../components/LogModal.svelte';
  import RepositoryModal from '../../components/RepositoryModal.svelte';
  import Snapshot from '../../components/Icon/Snapshot.svelte';

  let repos = [];
  let step = 0;
  let initialValues = null;
  let logId = null;

  const onRefresh = async () => {
    try {
      repos = await agent.get('/repositories');
      initialValues = repos.find((repo) => repo.id === initialValues.id);
    } catch (err) {
      // Pass
    }
  }
  
  const onRemove = async (repo) => {
    try {
      await agent.delete(`/repositories/${repo.id}`);
      await onRefresh()
    } catch (err) {
      console.log(err)
    }
  }

  $:if (initialValues === null) step = 0;
  $:values = Object.assign({}, initialValues) || step;

  onMount(onRefresh);
</script>

{#if initialValues}
  <RepositoryModal
    repository={initialValues}
    on:close={() => { initialValues = null; }}
    on:refresh={onRefresh}
  />
{:else if logId}
  <LogModal
    repositoryId={logId}
    on:close={() => { logId = null; }}
  />
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each repos as repo, i}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p class="truncate text-lg font-bold {repo.active ? '' :'text-red-500'}">#{repo.id} - {repo.name}</p>
      <p class="truncate text-sm font-light my-2">{dayjs(repo.createdAt).format('DD/MM/YY')}</p>

      <div class="flex flex-wrap gap-2 mt-auto pt-2 min-h-8 mb-1">
        <a href="/repositories/{repo.id}">
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
          on:click|preventDefault|stopPropagation={() => onRemove(repo)}
        >
          <Trash />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-orange-500 hover:bg-orange-600"
          title="Logs"
          on:click|preventDefault|stopPropagation={() => { logId = repo.id }}
        >
          <Logs />
        </button>
        <a href="/repositories/{repo.id}?tab=snapshots">
          <button
            class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-violet-500 hover:bg-violet-600"
            title="Snapshots"
          >
            <Snapshot />
          </button>
        </a>
      </div>
    </div>
  {/each}
  <div
    on:click={() => { initialValues = {}; }}
    class="text-3xl p-6 max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center flex justify-center cursor-pointer"
  >
    +
  </div>
</div>
