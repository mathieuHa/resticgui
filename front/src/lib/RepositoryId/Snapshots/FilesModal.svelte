<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import Modal from '$lib/Modal.svelte';
  import agent from '$utils/api';
  import Input from '$lib/Input.svelte';
  import Search from '$lib/Icon/Search.svelte';
  import Loader from '$lib/Loader.svelte';
  import Number from '$lib/Number.svelte';
  import File from '$lib/Icon/File.svelte';
  import Folder from '$lib/Icon/Folder.svelte';
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
      <table class="table-auto gap-5 w-full text-center">
        <thead>
          <tr>
            <th>Type</th>
            <th>Size</th>
            <th>Mod Time</th>
            <th class="text-left">Path</th>
            <th>Download</th>
          </tr>
        </thead>
        <tbody>
          {#each files as file}
            <tr>
              <td class="flex justify-center">
                {#if file.type == "dir" }
                  <Folder />
                {:else}
                  <File />
                {/if}
              </td>
              <td>
                <Number number={file.size}/>
              </td>
              <td>
                {file.mtime.split('.')[0].replace('T', ' ')}
              </td>
              <td class="text-left">
                {file.path}
              </td>
              <td class="flex justify-center">
                <a href={`http://localhost:3001/api/v1/repositories/${$page.params.id}/download/${snapshotId}?path=${file.path}`}>
                  <Download />
                </a>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {:else if pattern && pattern != ""}
      <p>
        No files in repository for pattern {pattern}
      </p>
    {:else}
      <p>
        No files in repository
      </p>
    {/if}
  {/await}
</Modal>