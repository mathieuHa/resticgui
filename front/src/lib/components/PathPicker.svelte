<script lang="ts">
  import CornerLeftUp from '@lucide/svelte/icons/corner-left-up';
  import File from '@lucide/svelte/icons/file';
  import Folder from '@lucide/svelte/icons/folder';
  import House from '@lucide/svelte/icons/house';

  import { api, ApiError, type Entry } from '#lib/api/client.ts';

  // Browses folders on the server, within the allowed roots.
  let {
    onpick,
    files = false,
    label = 'Use this folder',
  }: {
    onpick: (path: string) => void;
    /** also list files, and let them be picked */
    files?: boolean;
    label?: string;
  } = $props();

  let path = $state<string | undefined>(undefined);
  let entries = $state<Entry[]>([]);
  let error = $state('');
  let loading = $state(false);

  async function open(p: string | undefined) {
    loading = true;
    error = '';
    try {
      entries = await api.browse(p);
      path = p;
    } catch (err) {
      error = err instanceof ApiError ? err.message : String(err);
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    open(undefined);
  });

  const parent = $derived.by(() => {
    if (!path || path === '/') return undefined;
    const i = path.lastIndexOf('/');
    return i <= 0 ? '/' : path.slice(0, i);
  });
  const shown = $derived(files ? entries : entries.filter((e) => e.dir));
</script>

<div class="rounded-box border-base-300 border" data-testid="path-picker">
  <div class="bg-base-200 flex items-center gap-2 px-3 py-2 text-sm">
    <button type="button" class="btn btn-ghost btn-xs btn-square" onclick={() => open(undefined)} title="Allowed folders" aria-label="Allowed folders">
      <House class="size-4" aria-hidden="true" />
    </button>
    {#if path !== undefined}
      <button type="button" class="btn btn-ghost btn-xs btn-square" onclick={() => open(parent)} disabled={!parent} title="Parent folder" aria-label="Parent folder">
        <CornerLeftUp class="size-4" aria-hidden="true" />
      </button>
    {/if}
    <span class="truncate font-mono">{path ?? 'Allowed folders'}</span>
    {#if path !== undefined}
      <button type="button" class="btn btn-primary btn-xs ml-auto" onclick={() => path && onpick(path)}>{label}</button>
    {/if}
  </div>
  {#if error}
    <div class="text-error p-3 text-sm">{error}</div>
  {/if}
  <ul class="max-h-64 overflow-y-auto">
    {#if loading}
      <li class="p-3"><span class="loading loading-spinner loading-sm"></span></li>
    {:else if shown.length === 0}
      <li class="text-base-content/70 p-3 text-sm">Empty folder</li>
    {/if}
    {#each shown as entry (entry.path)}
      <li class="hover:bg-base-200 flex items-center gap-2 px-3 py-1 text-sm">
        {#if entry.dir}
          <button type="button" class="link link-hover flex flex-1 items-center gap-2 truncate text-left" onclick={() => open(entry.path)}>
            <Folder class="text-primary size-4 shrink-0" aria-hidden="true" />{entry.name}
          </button>
        {:else}
          <span class="flex flex-1 items-center gap-2 truncate"><File class="text-base-content/60 size-4 shrink-0" aria-hidden="true" />{entry.name}</span>
        {/if}
        <button type="button" class="btn btn-ghost btn-xs" onclick={() => onpick(entry.path)}>Pick</button>
      </li>
    {/each}
  </ul>
</div>
