<script lang="ts">
  import type { Paths } from '#lib/api/client.ts';

  import PathPicker from './PathPicker.svelte';

  // What to back up: folders and files to include, patterns to exclude.
  let { value = $bindable() }: { value: Paths } = $props();

  let typed = $state('');
  let excludes = $state(value.exclude.join('\n'));

  $effect(() => {
    value.exclude = excludes
      .split('\n')
      .map((l) => l.trim())
      .filter(Boolean);
  });

  function add(path: string) {
    const p = path.trim();
    if (p && !value.include.includes(p)) value.include.push(p);
    typed = '';
  }
</script>

<div class="flex flex-col gap-4">
  <div>
    <h4 class="mb-2 font-medium">Back up</h4>
    {#if value.include.length === 0}
      <p class="mb-2 text-sm opacity-60">Nothing yet. Pick folders below.</p>
    {/if}
    <ul class="mb-2 flex flex-col gap-1" data-testid="include-paths">
      {#each value.include as path, i (path)}
        <li class="bg-base-200 rounded-box flex items-center gap-2 px-3 py-1 font-mono text-sm">
          <span class="flex-1 truncate">{path}</span>
          <button type="button" class="btn btn-ghost btn-xs" aria-label="Remove {path}" onclick={() => value.include.splice(i, 1)}>✕</button>
        </li>
      {/each}
    </ul>
    <PathPicker onpick={add} files label="Add this folder" />
    <form
      class="join mt-2 w-full"
      onsubmit={(e) => {
        e.preventDefault();
        add(typed);
      }}
    >
      <input class="input input-sm join-item flex-1 font-mono" placeholder="/or/type/a/path" bind:value={typed} />
      <button class="btn btn-sm join-item" disabled={!typed.trim()}>Add</button>
    </form>
  </div>
  <label class="flex flex-col gap-1">
    <span class="font-medium">Exclude</span>
    <span class="text-xs opacity-60">One restic pattern per line, such as <span class="font-mono">*.tmp</span> or <span class="font-mono">/home/*/.cache</span></span>
    <textarea class="textarea w-full font-mono text-sm" rows="3" bind:value={excludes}></textarea>
  </label>
</div>
