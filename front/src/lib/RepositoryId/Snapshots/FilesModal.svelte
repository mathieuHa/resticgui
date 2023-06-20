<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import Loader from '$lib/Loader.svelte';
  import Number from '$lib/Number.svelte';
  import Download from '$lib/Icon/Download.svelte';
    import Modal from '$lib/Modal.svelte';

  export let snapshotId: string;
  let pattern: string | null = null;

  const dispatch = createEventDispatcher();
  const search = async () => agent.get(
    `/repositories/${$page.params.id}/snapshots/${snapshotId}/search${pattern ? `?pattern=${pattern}` : ''}`,
  );
  let searchPromise = search();
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<Modal on:close={() => dispatch('close')}>
  <div class="modal-box" on:click|preventDefault|stopPropagation={() => {}}>
    <h3 class="font-bold text-lg mb-4">Files snapshot #{snapshotId}</h3>
    <form class="form-control mb-4">
      <div class="input-group">
        <input
          type="text"
          placeholder="Search file"
          class="input input-bordered w-full"
          bind:value={pattern}
        />
        <button class="btn btn-square" type="submit" on:click={() => (searchPromise = search())}>
          <svg width="24" height="24" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </button>
      </div>
    </form>
  
    {#await searchPromise}
      <Loader size={8} />
    {:then files}
      {#if files && files.length > 0}
        <div class="overflow-y-auto relative flex flex-col gap-1 sm:text-center">
          <div class="sticky top-0 pb-1 sm:flex bg-base-200 border border-x-0 border-t-0 border-stone-500">
            <p class="flex-1 font-bold">Type</p>
            <p class="flex-1 font-bold">Size</p>
            <p class="flex-initial w-48 font-bold">Mod.Time</p>
            <p class="flex-initial w-64 text-left font-bold">Path</p>
            <p class="flex-1 font-bold">Download</p>
          </div>
          {#each files as file}
            <div class="sm:flex sm:text-center p-1 sm:p-0 border border-x-0 border-t-0 border-stone-500/75">
              <p class="flex-1" title={file.type == 'dir' ? 'folder' : 'file'}>
                {file.type == 'dir' ? '📂' : '📄'}
              </p>
              <p class="flex-1">
                {#if file.size}
                  <Number number={file.size}/>
                {:else}
                  <span>--</span>
                {/if}
              </p>
              <p class="flex-initial w-48">
                {file.mtime.split('.')[0].replace('T', ' ')}
              </p>
              <p class="flex-initial w-64 text-left">
                {file.path}
              </p>
              <div class="flex-1 flex sm:justify-center my-1 sm:mt-0">
                <a
                  class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-green-500 hover:bg-green-600"
                  href="/api/v1/repositories/{$page.params.id}/snapshots/{snapshotId}/download?path={file.path}"
                >
                  <Download />
                </a>
              </div>
            </div>
          {/each}
        </div>
      {:else if pattern && pattern != ""}
        <p>
          No files in repository for pattern {pattern}
        </p>
      {:else}
        <p>
          No files in repository
        </p>
      {/if}
    {:catch err}
      <pre>{err}</pre>
    {/await}
  </div>
</Modal>