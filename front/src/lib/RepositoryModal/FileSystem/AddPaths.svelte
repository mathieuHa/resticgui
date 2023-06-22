<script lang="ts">
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
      <div class="flex items-center leading-4 {(paths.includes(`!${path}/${key}`) && 'text-warning') || ((included || paths.includes(`${path}/${key}`)) && 'text-green-500') || ''}">
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
        <svg width="12px" height="12px" fill="currentColor" viewBox="0 0 330 330">
          <path d="M325.606,229.393l-150.004-150C172.79,76.58,168.974,75,164.996,75c-3.979,0-7.794,1.581-10.607,4.394l-149.996,150c-5.858,5.858-5.858,15.355,0,21.213c5.857,5.857,15.355,5.858,21.213,0l139.39-139.393l139.397,139.393C307.322,253.536,311.161,255,315,255c3.839,0,7.678-1.464,10.607-4.394C331.464,244.748,331.464,235.251,325.606,229.393z"/>
        </svg>
        </button>
        <p class="flex-1 h-5 leading-4 truncate overflow-hidden {(paths.includes(`!${path}/${key}`) && 'text-warning') || ((included || paths.includes(`${path}/${key}`)) && 'text-green-500') || paths.find((p) => p.startsWith(`${path}/${key}`)) && 'text-blue-500' || ''}">
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

