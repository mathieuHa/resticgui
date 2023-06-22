<script lang="ts">
  import dayjs from 'dayjs';

  import agent from '$utils/api';
  import Eye from '$lib/Icon/Eye.svelte';
  import Logs from '$lib/Icon/Logs.svelte';
  import Pen from '$lib/Icon/Pen.svelte';
  import Trash from '$lib/Icon/Trash.svelte';
  import RepositoryModal from '$lib/RepositoryModal/RepositoryModal.svelte';
  import Snapshot from '$lib/Icon/Snapshot.svelte';
    import Modal from '$lib/Modal.svelte';

  let repos = [];
  let step = 0;
  let initialValues: any = null;
  let removeId: number | null = null;

  const onRefresh = async () => {
    repos = await agent.get('/repositories');
    initialValues = repos.find((repo: any) => repo.id === initialValues?.id);
    return repos;
  };
  let reposPromise = onRefresh()

  const onRemove = async (repoId: number) => {
    try {
      await agent.delete(`/repositories/${repoId}`);
      reposPromise = onRefresh();
      removeId = null;
    } catch (err) {
      console.log(err);
    }
  };

  $: if (initialValues === null) step = 0;
  $: values = Object.assign({}, initialValues) || step;

</script>

{#if removeId}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <Modal on:close={() => (removeId = null)}>
    <div class="modal-box">
      <h3 class="font-bold text-lg">Warning</h3>
      <p class="py-4">Do you want to delete repository #{removeId} ?</p>
      <div class="modal-action">
        <button class="btn btn-info" on:click={() => (removeId = null)}>No</button>
        <button class="btn btn-error" on:click={() => removeId && onRemove(removeId)}>Yes</button>
      </div>
    </div>
  </Modal>
{:else if initialValues}
  <RepositoryModal
    repository={initialValues}
    on:close={() => (initialValues = null)}
    on:refresh={() => (reposPromise = onRefresh())}
  />
{/if}

<div class="w-full flex flex-col gap-5 mb-6">
  {#await reposPromise}
    <div class="p-3 flex flex-col bg-base-100 rounded-xl shadow-lg h-24 animate-pulse">
      <div class="flex justify-between gap-2">
        <div class="w-36 bg-base-50 h-7 rounded-md " />
        <div class="w-24 bg-base-50 h-7 rounded-md " />
      </div>
      <div class="flex flex-wrap gap-2 mt-5 mb-1">
        <div class="rounded h-6 w-6 p-1 transition bg-yellow-500" />
        <div class="rounded h-6 w-6 p-1 transition bg-blue-500" />
        <div class="rounded h-6 w-6 p-1 transition bg-red-500" />
        <div class="rounded h-6 w-6 p-1 transition bg-orange-500" />
        <div class="rounded h-6 w-6 p-1 transition bg-purple-500" />
      </div>
    </div>
    <div class="h-12 bg-base-100 rounded-xl shadow-lg animate-pulse" />
  {:then repos}
    {#each repos as repo, i}
      <div class="p-3 flex flex-col bg-base-100 rounded-xl shadow-lg">
        <div class="flex flex-wrap gap-2">
          <p class="text-lg font-bold leading-7 whitespace-nowrap {repo.active ? '' :'text-error'}">#{repo.id} - {repo.name}</p>
          <p class="flex-1 text-sm font-light leading-7 text-right whitespace-nowrap">{dayjs(repo.createdAt).format('DD/MM/YYYY HH:mm ')}</p>
        </div>

        <div class="flex flex-wrap justify-between mt-5 mb-1">
          <div class="flex flex-wrap gap-2 min-w-[200px]">
            <a href="/ui/{repo.id}">
              <div
                class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-yellow-500 hover:bg-yellow-600"
                title="See"
              >
                <Eye />
              </div>
            </a>
            <button
              class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
              title="Update"
              on:click|preventDefault|stopPropagation={() => {
                initialValues = Object.assign({}, repo);
              }}
            >
              <Pen />
            </button>
            <button
              class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-red-500 hover:bg-red-600"
              title="Remove"
              on:click|preventDefault|stopPropagation={() => (removeId = repo.id)}
            >
              <Trash />
            </button>
            <a href="/ui/{repo.id}?tab=logs">
              <div
                class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-orange-500 hover:bg-orange-600"
                title="Logs"
              >
                <Logs />
              </div>
            </a>
            <a href="/ui/{repo.id}?tab=snapshots">
              <div
                class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-violet-500 hover:bg-violet-600"
                title="Snapshots"
              >
                <Snapshot />
              </div>
            </a>
          </div>
          <span class="text-{values.running ? 'green' : 'red'}-500 flex-1 pl-1 whitespace-nowrap text-right ">
            {values.running ? 'running' : 'not running'}
          </span>
        </div>
      </div>
    {/each}
    <button
      on:click={() => { initialValues = {}; }}
      class="btn btn-ghost h-12 bg-base-100 rounded-xl shadow-lg normal-case text-base"
    >
      Add a repository
    </button>
  {/await}
</div>
