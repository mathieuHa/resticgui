<script lang="ts">
  import dayjs from 'dayjs';
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import Loader from '$lib/Loader.svelte';
  import Number from '$lib/Number.svelte';
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
  <div class="modal-box">
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
        <div class="overflow-x-auto w-full">
          <table class="table table-sm table-zebra table-pin-rows w-full">
            <!-- head -->
            <thead>
              <tr>
                <th class="pl-0">Type</th>
                <th class="pl-0">Size</th>
                <th class="pl-0">Mod.Time</th>
                <th class="pl-0">Path</th>
                <th class="px-0 flex justify-end">Download</th>
              </tr>
            </thead>
            <tbody>
              {#each files as file}
                <tr>
                  <th class="pl-0">{file.type == 'dir' ? '📂' : '📄'}</th>
                  <td class="pl-0">
                    {#if file.size}
                      <Number number={file.size}/>
                    {:else}
                      <span>--</span>
                    {/if}
                  </td>
                  <td class="pl-0">{dayjs(file.mtime).format('DD/MM/YYYY HH:mm')}</td>
                  <td class="pl-0">{file.path}</td>
                  <td class="px-0 flex justify-end">
                    <a
                      class="btn btn-circle btn-xs btn btn-success"
                      href="/api/v1/repositories/{$page.params.id}/snapshots/{snapshotId}/download?path={file.path}"
                      download={file.path.split('/').pop()}
                    >
                      <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" stroke="currentColor">
                        <path d="M.5 9.9a.5.5 0 0 1 .5.5v2.5a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-2.5a.5.5 0 0 1 1 0v2.5a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2v-2.5a.5.5 0 0 1 .5-.5z"/>
                        <path d="M7.646 11.854a.5.5 0 0 0 .708 0l3-3a.5.5 0 0 0-.708-.708L8.5 10.293V1.5a.5.5 0 0 0-1 0v8.793L5.354 8.146a.5.5 0 1 0-.708.708l3 3z"/>
                      </svg>
                    </a>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
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