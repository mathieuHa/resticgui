<script lang="ts">
  import AddPaths from './AddPaths.svelte';
  import Arrow from '../../Icon/Arrow.svelte';

  export let tree = {};
  export let path = '';
  export let paths: string[] = [];
  export let onTogglePaths: (path: string, include: boolean) => void;
  export let onDirClick: (root: string, fetched: boolean, visible: boolean) => void;


  const isIncluded = (pathsCopy: string[]) => {
    let deepInclude = { name: '', included: false };
    for (let i = 0; i < pathsCopy.length; i += 1) {
      if (deepInclude.name.length < pathsCopy[i].length
        && (`${path}/`.startsWith(`${pathsCopy[i]}/`) || `!${path}/`.startsWith(`${pathsCopy[i]}/`))) {
        deepInclude = pathsCopy[i][0] !== '!'
          ? { included: true, name: pathsCopy[i] }
          : { included: false, name: pathsCopy[i].slice(1) };
      }
    }
    return deepInclude.included;
  }
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
          on:click|preventDefault|stopPropagation={async () => onDirClick(`${path}/${key}/`, tree[key].__fetched, tree[key].__visible)}
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
        <AddPaths path={`${path}/${key}`} tree={tree[key]} {paths} {onTogglePaths} {onDirClick} />
      </div>
    {/if}
  {/each}
</div>

