
<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { page } from '$app/stores';
  import Modal from '$lib/Modal.svelte';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';
  import Download from '$lib/Icon/Download.svelte';
  import File from '$lib/Icon/File.svelte';
  import Folder from '$lib/Icon/Folder.svelte';
  import Number from '$lib/Number.svelte';

  export let snapshotId: string | null = null;
  let files: any[] = []

  const dispatch = createEventDispatcher();
  onMount(
    async () => {
      try {
        files = await agent.get(`/repositories/${$page.params.id}/ls/${snapshotId}`);
        toasts.addToast(`List files of ${snapshotId} received`);
      // dispatch('close');
      } catch (err) {
        toasts.addToast(err as string, 'failure');
      }
    }
  )
  
</script>

<Modal title="Files in snapshot #{snapshotId}" on:close={() => dispatch('close')}>
  <div>
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
            <td><a href={`http://localhost:3001/api/v1/repositories/${$page.params.id}/download/${snapshotId}?path=${f.path}`}><Download /></a></td>
          </tr>
          {/each}
        </tbody>
      </table>
    </div>
    {:else}
      No files in repository
    {/if}
  </div>
</Modal>