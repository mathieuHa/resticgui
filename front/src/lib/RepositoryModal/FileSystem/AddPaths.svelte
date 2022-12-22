<script lang="ts">
  import Arrow from '$lib/Icon/Arrow.svelte';
  import type { Tree } from '$utils/types';

  export let tree: Tree;
  export let path = '';
  export let paths: string[] = [];
  export let onTogglePaths: (path: string, include: boolean) => void;
  export let onDirClick: (root: string, fetched: boolean, visible: boolean) => Promise<void>;


  const isIncluded = (pathsCopy: string[]) => {
    let deepInclude = { name: '', included: false };
    pathsCopy.forEach((pathCopy) => {
      if (deepInclude.name.length < pathCopy.length
        && (`${path}/`.startsWith(`${pathCopy}/`) || `!${path}/`.startsWith(`${pathCopy}/`))) {
        deepInclude = pathCopy[0] !== '!'
          ? { included: true, name: pathCopy }
          : { included: false, name: pathCopy.slice(1) };
      }
    });
    return deepInclude.included;
  }
  let loading = false;
  let included = false;
  $:included = isIncluded(paths)
</script>

<div>
  {#each Object.entries(tree) as [key, value]}
    {#if typeof value === 'string' && key !== '__absPath'}
      <div class="flex items-center leading-4 {(paths.includes(`!${path}/${key}`) && 'text-orange-500') || ((included || paths.includes(`${path}/${key}`)) && 'text-green-500') || ''}">
        <button
          title={value}
          class="cursor-pointer h-5 truncate overflow-hidden"
          on:click|preventDefault|stopPropagation={() => onTogglePaths(`${path}/${key}`, !included)}
        >
          {key}
        </button>
      </div>
    {:else if typeof value === 'object'}
      <div class="flex items-center">
        <button
          class="w-3 h-3 mr-2 cursor-pointer {tree[key].__visible ? 'rotate-180' : 'rotate-90'} transition"
          on:click|preventDefault|stopPropagation={async () => {
            loading = true;
            await onDirClick(`${path}/${key}/`, tree[key].__fetched, tree[key].__visible);
            loading = false;
          }}
        >
          <Arrow />
        </button>
        <p class="flex-1 h-5 leading-4 truncate overflow-hidden {(paths.includes(`!${path}/${key}`) && 'text-orange-500') || ((included || paths.includes(`${path}/${key}`)) && 'text-green-500') || paths.find((p) => p.startsWith(`${path}/${key}`)) && 'text-blue-500' || ''}">
          <button
            class="cursor-pointer"
            on:click|preventDefault|stopPropagation={() => onTogglePaths(`${path}/${key}`, !included)}
          >
            {key}
          </button>
        </p>
      </div>
      <div class={`pl-5 ${tree[key].__visible ? '' : 'hidden'}`}>
        <svelte:self path={`${path}/${key}`} tree={tree[key]} {paths} {onTogglePaths} {onDirClick} />
      </div>
    {/if}
  {/each}
</div>

