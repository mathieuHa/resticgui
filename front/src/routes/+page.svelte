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
    <div class="p-3 flex flex-col bg-base-100 rounded-xl shadow-md animate-pulse">
      <div><div class="w-36 h-5 my-1 bg-neutral-content rounded-xl" /></div>
      <div class="flex flex-wrap gap-2 mt-5 mb-1">
        <div class="btn btn-square btn-sm btn-warning" />
        <div class="btn btn-square btn-sm btn-info" />
        <div class="btn btn-square btn-sm btn-error" />
        <div class="btn btn-square btn-sm btn-success" />
        <div class="btn btn-square btn-sm btn-primary" />
      </div>
    </div>
    <div class="btn btn-ghost bg-base-100 hover:bg-base-200 rounded-xl shadow-md animate-pulse" />
  {:then repos}
    {#each repos as repo, i}
      <div class="p-3 flex flex-col bg-base-100 rounded-xl shadow-md">
        <div class="flex flex-wrap h-7 items-end gap-2">
          <p class="text-lg font-bold whitespace-nowrap {repo.active ? '' :'text-error'}">#{repo.id} - {repo.name}</p>
          <p class="flex-1 text-sm font-light text-right whitespace-nowrap">
            {dayjs(repo.createdAt).format('DD/MM/YYYY HH:mm')}
          </p>
        </div>

        <div class="flex flex-wrap justify-between mt-5 mb-1">
          <div class="flex flex-wrap gap-2">
            <a href="/ui/{repo.id}" class="btn btn-square btn-sm btn-warning" title="See">
              <Eye />
            </a>
            <button
              class="btn btn-square btn-sm btn-info"
              title="Update"
              on:click|stopPropagation={() => (initialValues = Object.assign({}, repo))}
            >
              <Pen />
            </button>
            <button
              class="btn btn-square btn-sm btn-error"
              title="Remove"
              on:click|stopPropagation={() => (removeId = repo.id)}
            >
              <Trash />
            </button>
            <a href="/ui/{repo.id}?tab=logs" class="btn btn-square btn-sm btn-success" title="Logs">
              <Logs />
            </a>
            <a href="/ui/{repo.id}?tab=snapshots" class="btn btn-square btn-sm btn-primary" title="Snapshots">
              <Snapshot />
            </a>
          </div>
          <span class="{values.running ? '' : 'text-red-500'} flex-1 pl-1 whitespace-nowrap text-right ">
            {values.running ? 'running' : 'not running'}
          </span>
        </div>
      </div>
    {/each}
    <button
      on:click={() => { initialValues = {}; }}
      class="btn btn-ghost bg-base-100 hover:bg-base-200 rounded-xl shadow-md normal-case text-base"
    >
      Add a repository
    </button>
  {/await}
</div>
