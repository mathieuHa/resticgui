<script lang="ts">
  import { api, ApiError, type FileNode, type FindMatch, type Snapshot } from '#lib/api/client.ts';
  import { bytes, dateTime } from '#lib/format.ts';

  import RestoreDialog from './RestoreDialog.svelte';

  // The files of one snapshot: browse, search, download and restore.
  let { repositoryId, snapshot }: { repositoryId: number; snapshot: Snapshot } = $props();

  let dir = $state('/');
  let nodes = $state<FileNode[]>([]);
  let error = $state('');
  let loading = $state(false);
  let pattern = $state('');
  let matches = $state<FindMatch[] | null>(null);
  let restoreOpen = $state(false);
  let restoreInclude = $state<string[]>([]);

  async function open(path: string) {
    loading = true;
    error = '';
    matches = null;
    try {
      nodes = await api.files(repositoryId, snapshot.id, path);
      dir = path;
    } catch (err) {
      error = err instanceof ApiError ? err.message : String(err);
    } finally {
      loading = false;
    }
  }

  async function find(e: SubmitEvent) {
    e.preventDefault();
    if (!pattern.trim()) return;
    loading = true;
    error = '';
    try {
      matches = await api.find(repositoryId, snapshot.id, pattern.trim());
    } catch (err) {
      error = err instanceof ApiError ? err.message : String(err);
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    open('/');
  });

  const crumbs = $derived.by(() => {
    const parts = dir.split('/').filter(Boolean);
    return [{ name: '/', path: '/' }, ...parts.map((name, i) => ({ name, path: '/' + parts.slice(0, i + 1).join('/') }))];
  });

  function restore(include: string[]) {
    restoreInclude = include;
    restoreOpen = true;
  }
</script>

<div class="flex flex-col gap-3" data-testid="snapshot-browser">
  <div class="flex flex-wrap items-center gap-2">
    <span class="font-mono text-sm">{snapshot.short_id}</span>
    <span class="text-sm opacity-70">{dateTime(snapshot.time)} on {snapshot.hostname}</span>
    <div class="ml-auto flex gap-2">
      <a class="btn btn-sm" href={api.downloadURL(repositoryId, snapshot.id, '/', 'zip')} download>Download all</a>
      <button class="btn btn-sm btn-primary" onclick={() => restore([])}>Restore all…</button>
    </div>
  </div>

  <form class="join w-full" onsubmit={find}>
    <input class="input input-sm join-item flex-1" placeholder="Find files by name, such as *.jpg" bind:value={pattern} />
    <button class="btn btn-sm join-item">Find</button>
    {#if matches}
      <button type="button" class="btn btn-sm join-item" onclick={() => open(dir)}>Back to folders</button>
    {/if}
  </form>

  {#if error}<div class="alert alert-error text-sm">{error}</div>{/if}

  {#if matches}
    <p class="text-sm opacity-70">{matches.length} found</p>
    <ul class="max-h-96 overflow-y-auto text-sm">
      {#each matches as m (m.path)}
        <li class="hover:bg-base-200 flex items-center gap-2 px-2 py-1">
          <span class="flex-1 truncate font-mono">{m.path}</span>
          <span class="opacity-60">{m.type === 'dir' ? '' : bytes(m.size)}</span>
          <a class="btn btn-ghost btn-xs" href={api.downloadURL(repositoryId, snapshot.id, m.path, m.type === 'dir' ? 'zip' : 'none')} download>Download</a>
          <button class="btn btn-ghost btn-xs" onclick={() => restore([m.path])}>Restore…</button>
        </li>
      {/each}
    </ul>
  {:else}
    <div class="breadcrumbs text-sm">
      <ul>
        {#each crumbs as c (c.path)}
          <li><button class="link link-hover font-mono" onclick={() => open(c.path)}>{c.name}</button></li>
        {/each}
      </ul>
    </div>
    <div class="max-h-[28rem] overflow-y-auto">
      <table class="table-sm table">
        <tbody>
          {#if loading}
            <tr><td><span class="loading loading-spinner loading-sm"></span></td></tr>
          {/if}
          {#each nodes as node (node.path)}
            <tr class="hover:bg-base-200">
              <td class="w-full">
                {#if node.type === 'dir'}
                  <button class="link link-hover font-mono" onclick={() => open(node.path)}>📁 {node.name}</button>
                {:else}
                  <span class="font-mono">{node.type === 'symlink' ? '🔗' : '📄'} {node.name}</span>
                {/if}
              </td>
              <td class="whitespace-nowrap opacity-60">{node.type === 'dir' ? '' : bytes(node.size)}</td>
              <td class="whitespace-nowrap opacity-60">{dateTime(node.mtime)}</td>
              <td class="whitespace-nowrap">
                <a
                  class="btn btn-ghost btn-xs"
                  href={api.downloadURL(repositoryId, snapshot.id, node.path, node.type === 'dir' ? 'zip' : 'none')}
                  download>Download</a
                >
                <button class="btn btn-ghost btn-xs" onclick={() => restore([node.path])}>Restore…</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<RestoreDialog bind:open={restoreOpen} {repositoryId} {snapshot} include={restoreInclude} />
