<script lang="ts">
  import { page } from '$app/stores';
  import Download from '$lib/Icon/Download.svelte';
  import File from '$lib/Icon/File.svelte';
  import Folder from '$lib/Icon/Folder.svelte';
  import Number from '$lib/Number.svelte';

  export let files = [];
  export let pattern: string | null = null;
  export let snapshotId: string;
</script>

{#if files && files.length > 0}
  <div class="grid gap-2">
    <table class="table-auto text-left">
      <thead>
        <tr>
          <th>Type</th>
          <th>Size</th>
          <th>Mod Time</th>
          <th>Path</th>
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
          <td>
            {file.path}
          </td>
          <td>
            <a class="inline mx-auto" href={`http://localhost:3001/api/v1/repositories/${$page.params.id}/download/${snapshotId}?path=${file.path}`}>
              <Download />
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