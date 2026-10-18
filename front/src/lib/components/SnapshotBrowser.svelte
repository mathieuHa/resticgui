<script lang="ts">
  import ArchiveRestore from '@lucide/svelte/icons/archive-restore';
  import Download from '@lucide/svelte/icons/download';
  import File from '@lucide/svelte/icons/file';
  import Folder from '@lucide/svelte/icons/folder';
  import Link from '@lucide/svelte/icons/link';
  import Search from '@lucide/svelte/icons/search';
  import { tick } from 'svelte';

  import { api, ApiError, type FileNode, type FindMatch, type Snapshot } from '#lib/api/client.ts';
  import { bytes, dateTime } from '#lib/format.ts';
  import { moveSelection, typing } from '#lib/keys.ts';

  import RestoreDialog from './RestoreDialog.svelte';

  // The files of one snapshot: browse, search, download and restore, with
  // the mouse or the keyboard.
  let { repositoryId, snapshot }: { repositoryId: number; snapshot: Snapshot } = $props();

  let dir = $state('/');
  let nodes = $state<FileNode[]>([]);
  let error = $state('');
  let loading = $state(false);
  let pattern = $state('');
  let matches = $state<FindMatch[] | null>(null);
  let restoreOpen = $state(false);
  let restoreInclude = $state<string[]>([]);
  let selected = $state(-1);
  let list: HTMLElement | undefined = $state();
  let search: HTMLInputElement | undefined = $state();

  type Item = { path: string; name: string; type: string; size?: number; mtime?: string };
  const items = $derived<Item[]>(matches ? matches.map((m) => ({ ...m, name: m.path })) : nodes);

  async function open(path: string, focusChild?: string) {
    loading = true;
    error = '';
    matches = null;
    try {
      nodes = await api.files(repositoryId, snapshot.id, path);
      dir = path;
      selected = focusChild ? Math.max(0, nodes.findIndex((n) => n.path === focusChild)) : nodes.length ? 0 : -1;
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
      selected = matches.length ? 0 : -1;
      list?.focus();
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
  const parent = $derived(dir === '/' ? null : dir.slice(0, dir.lastIndexOf('/')) || '/');

  function restore(include: string[]) {
    restoreInclude = include;
    restoreOpen = true;
  }

  const downloadURL = (item: Item) => api.downloadURL(repositoryId, snapshot.id, item.path, item.type === 'dir' ? 'zip' : 'none');

  async function select(i: number) {
    selected = i;
    await tick();
    list?.querySelector(`[data-index="${i}"]`)?.scrollIntoView({ block: 'nearest' });
  }

  function onkeydown(e: KeyboardEvent) {
    if (typing(e) || e.ctrlKey || e.metaKey || e.altKey) return;
    const item = items[selected];
    const move = moveSelection(e.key, selected, items.length);
    if (move !== null) {
      e.preventDefault();
      select(move);
      return;
    }
    switch (e.key) {
      case 'Enter':
      case 'ArrowRight':
        if (item?.type === 'dir') {
          e.preventDefault();
          open(item.path);
        }
        break;
      case 'Backspace':
      case 'ArrowLeft':
        if (!matches && parent) {
          e.preventDefault();
          open(parent, dir);
        } else if (matches && e.key === 'Backspace') {
          e.preventDefault();
          open(dir);
        }
        break;
      case '/':
        e.preventDefault();
        search?.focus();
        break;
      case 'r':
        if (item) restore([item.path]);
        break;
      case 'd':
        if (item) list?.querySelector<HTMLAnchorElement>(`[data-index="${selected}"] a[download]`)?.click();
        break;
    }
  }
</script>

<div class="flex flex-col gap-3" data-testid="snapshot-browser">
  <div class="flex flex-wrap items-center gap-2">
    <span class="badge badge-ghost font-mono">{snapshot.short_id}</span>
    <span class="text-base-content/70 text-sm"><time datetime={snapshot.time}>{dateTime(snapshot.time)}</time> on {snapshot.hostname}</span>
    <div class="ml-auto flex gap-2">
      <a class="btn btn-sm" href={api.downloadURL(repositoryId, snapshot.id, '/', 'zip')} download><Download class="size-4" aria-hidden="true" />Download all</a>
      <button class="btn btn-sm btn-primary" onclick={() => restore([])}><ArchiveRestore class="size-4" aria-hidden="true" />Restore all…</button>
    </div>
  </div>

  <form class="join w-full" onsubmit={find}>
    <label class="input input-sm join-item flex-1">
      <Search class="size-4 text-base-content/70" aria-hidden="true" />
      <input
        bind:this={search}
        class="grow"
        placeholder="Find files by name, such as *.jpg"
        aria-label="Find files by name"
        bind:value={pattern}
        onkeydown={(e) => {
          if (e.key === 'Escape') list?.focus();
        }}
      />
      <kbd class="kbd kbd-xs hidden sm:inline">/</kbd>
    </label>
    <button class="btn btn-sm join-item">Find</button>
    {#if matches}
      <button type="button" class="btn btn-sm join-item" onclick={() => open(dir)}>Back to folders</button>
    {/if}
  </form>

  {#if error}<div class="alert alert-error text-sm" role="alert">{error}</div>{/if}

  {#if matches}
    <p class="text-base-content/70 text-sm" role="status">{matches.length} found</p>
  {:else}
    <nav class="breadcrumbs py-0 text-sm" aria-label="Folder">
      <ul>
        {#each crumbs as c (c.path)}
          <li><button class="link link-hover font-mono" onclick={() => open(c.path)}>{c.name}</button></li>
        {/each}
      </ul>
    </nav>
  {/if}

  <!-- the list takes the keyboard: arrows to move, Enter to open, r to restore, d to download -->
  <!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions -->
  <div
    bind:this={list}
    class="border-base-300 focus-visible:outline-primary max-h-[28rem] overflow-y-auto rounded-lg border outline-offset-2 focus-visible:outline-2"
    tabindex="0"
    role="region"
    aria-label="Files in {matches ? 'the search results' : dir}"
    {onkeydown}
  >
    <table class="table-sm table">
      <thead class="bg-base-100 sticky top-0 z-[1]">
        <tr><th class="w-full">Name</th><th class="text-right">Size</th><th>Modified</th><th><span class="sr-only">Actions</span></th></tr>
      </thead>
      <tbody>
        {#if loading}
          {#each [1, 2, 3] as i (i)}<tr><td colspan="4"><div class="skeleton h-5"></div></td></tr>{/each}
        {:else if items.length === 0}
          <tr><td colspan="4" class="text-base-content/70">{matches ? 'Nothing matches.' : 'Empty folder.'}</td></tr>
        {/if}
        {#each loading ? [] : items as item, i (item.path)}
          {@const Icon = item.type === 'dir' ? Folder : item.type === 'symlink' ? Link : File}
          <tr
            data-index={i}
            aria-selected={i === selected}
            class="cursor-default {i === selected ? 'bg-primary/10' : 'hover:bg-base-200'}"
            onclick={() => (selected = i)}
          >
            <td class="max-w-0">
              {#if item.type === 'dir'}
                <button class="link link-hover flex max-w-full items-center gap-2 font-mono" onclick={() => open(item.path)}>
                  <Icon class="text-primary size-4 shrink-0" aria-hidden="true" /><span class="truncate">{item.name}</span>
                </button>
              {:else}
                <span class="flex items-center gap-2 font-mono">
                  <Icon class="text-base-content/60 size-4 shrink-0" aria-hidden="true" /><span class="truncate">{item.name}</span>
                </span>
              {/if}
            </td>
            <td class="tabular text-base-content/70 text-right whitespace-nowrap">{item.type === 'dir' ? '' : bytes(item.size)}</td>
            <td class="text-base-content/70 whitespace-nowrap">{#if item.mtime}<time datetime={item.mtime}>{dateTime(item.mtime)}</time>{/if}</td>
            <td class="whitespace-nowrap">
              <a class="btn btn-ghost btn-xs" href={downloadURL(item)} download>Download</a>
              <button class="btn btn-ghost btn-xs" onclick={() => restore([item.path])}>Restore…</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <p class="text-base-content/70 hidden flex-wrap gap-x-3 gap-y-1 text-xs sm:flex" aria-hidden="true">
    <span><kbd class="kbd kbd-xs">↑</kbd> <kbd class="kbd kbd-xs">↓</kbd> move</span>
    <span><kbd class="kbd kbd-xs">Enter</kbd> open</span>
    <span><kbd class="kbd kbd-xs">Backspace</kbd> up</span>
    <span><kbd class="kbd kbd-xs">/</kbd> find</span>
    <span><kbd class="kbd kbd-xs">r</kbd> restore</span>
    <span><kbd class="kbd kbd-xs">d</kbd> download</span>
  </p>
</div>

<RestoreDialog bind:open={restoreOpen} {repositoryId} {snapshot} include={restoreInclude} />
