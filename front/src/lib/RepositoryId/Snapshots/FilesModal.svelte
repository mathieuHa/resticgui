<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import Modal from '$lib/Modal.svelte';
  import agent from '$utils/api';
  import Input from '$lib/Input.svelte';
  import Search from '$lib/Icon/Search.svelte';
  import Loader from '$lib/Loader.svelte';
  import Number from '$lib/Number.svelte';
  import Download from '$lib/Icon/Download.svelte';

  export let snapshotId: string;
  let pattern: string | null = null;

  const dispatch = createEventDispatcher();
  const search = async () => agent.get(
    `/repositories/${$page.params.id}/search/${snapshotId}${pattern ? `?pattern=${pattern}` : ''}`,
  );
  let searchPromise = search();
</script>

<Modal title="Files found in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
  <div class="flex gap-2">
    <div class="flex-1">
      <Input
        secondary
        id="search"
        label="search file"
        type="text"
        bind:value={pattern}
        on:change={() => (searchPromise = search())}
      />
    </div>
    <button
      class="mt-auto mb-2 rounded h-8 w-8 text-gray-900 dark:text-gray-100 p-2 transition bg-blue-500 hover:bg-blue-600"
      on:click={() => (searchPromise = search())}
    >
      <Search />
    </button>
  </div>
  
  {#await searchPromise}
    <Loader size={8} />
  {:then files}
    {#if files && files.length > 0}
      <div class="relative overflow-y-scroll flex flex-col gap-1 custom-max-height pr-1 sm:text-center">
        <div class="sticky top-0 sm:flex bg-gray-200 dark:bg-gray-700 border border-x-0 border-t-0 border-stone-500">
          <p class="flex-1 font-bold">Type</p>
          <p class="flex-1 font-bold">Size</p>
          <p class="flex-initial w-48 font-bold">Mod.Time</p>
          <p class="flex-initial w-64 text-left font-bold">Path</p>
          <p class="flex-1 font-bold">Download</p>
        </div>
        {#each files as file}
          <div class="sm:flex sm:text-center p-1 sm:p-0 border border-x-0 border-t-0 border-stone-500/75">
            <p class="flex-1">
              {file.type == "dir" ? '📂' : '📄'}
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
                href={`http://localhost:3001/api/v1/repositories/${$page.params.id}/download/${snapshotId}?path=${file.path}`}
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
</Modal>

<style>
  .custom-max-height {
    max-height: calc(100vh - 200px);
  }
</style>