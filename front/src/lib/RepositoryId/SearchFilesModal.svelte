<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Modal from "$lib/Modal.svelte";
  import Download from '../Icon/Download.svelte';
  import File from '../Icon/File.svelte';
  import Folder from '../Icon/Folder.svelte';
  import {page} from '$app/stores';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';
  import { each, mount_component, onMount } from 'svelte/internal';
  import Number from './Number.svelte';
  import Input from '../Input.svelte';
  import Search from '../Icon/Search.svelte';

  export let searchId: string | null = null;
  let files: any[] = []
  let pattern: string | null = null;

  const dispatch = createEventDispatcher();
  const search = async () => {
      try {
        files = await agent.get(`/repositories/${$page.params.id}/search/${searchId}?pattern=${pattern}`);
        toasts.addToast(`Search pattern of filenames of ${searchId} received`);
      // dispatch('close');
      } catch (err) {
        toasts.addToast(err as string, 'failure');
      }
    }
  
</script>

<Modal title="Files found in snapshot #{searchId}" on:close={() => dispatch('close')}>
  <div>
    <div class="grid grid-flow-col">
      <div class="col">
        <Input
          id="search"
          label="search file"
          type="text"
          bind:value={pattern}
        />
      </div>
      <div class="col">
        <button
          class="bg-blue-500 hover:bg-blue-700 text-white py-2 px-2 rounded transition text-right align-middle"
          on:click={() => search()}>
          <Search />
        </button>
      </div>
    </div>
    
    
    {#if files && files.length > 0}
    <div class="grid">
      <table class="table-auto">
        <thead>
          <tr>
            <th class="text-left">Type</th>
            <th class="text-left">Size</th>
            <th class="text-left">Mod Time</th>
            <th class="text-left">Path</th>
            <th class="text-left">Download</th>
          </tr>
        </thead>
        <tbody>
          {#each files as f}
          <tr>
            <td>
              {#if f.type == "dir" }
                <Folder />
              {:else}
                <File />
              {/if}
            </td>
            <td><Number number={f.size}/></td>
            <td>{f.mtime.split('.')[0].replace('T', ' ')}</td>
            <td class="text-left">{f.path}</td>
            <td><a href={`http://localhost:3001/api/v1/repositories/${$page.params.id}/download/${searchId}?path=${f.path}`}><Download /></a></td>
          </tr>
          {/each}
        </tbody>
      </table>
    </div>
    {:else if pattern && pattern != ""}
      No files in repository for pattern {pattern}
    {/if}
  </div>
</Modal>