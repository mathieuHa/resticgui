<script>
  import AddPaths from './AddPaths.svelte';
  import Arrow from './icon/Arrow.svelte';

  export let tree = {};
  export let path = '';
  export let paths = [];
  export let onTogglePaths;
  export let onDirClick;

  const isIncluded = (pathsCopy) => {
    let i = 0;
    for (i; i < pathsCopy.length && !`${path}/`.startsWith(`${pathsCopy[i]}/`); i += 1);
    return i !== pathsCopy.length;
  }
  let included;
  $:included = isIncluded(paths)
</script>

<div>
  {#each Object.entries(tree) as [key, value]}
    {#if typeof value === 'string' && key !== '__absPath'}
      <div class="flex items-center leading-4 {included || paths.includes(`${path}/${key}`) ? 'text-green-500' : ''}">
        <span title={value} class="cursor-pointer h-5" on:click={() => onTogglePaths([`${path}/${key}`])}>{key}</span>
      </div>
    {:else if typeof value === 'boolean'}
    {:else if typeof value === 'object'}
      <div class="flex items-center">
        <div class="w-3 h-3 mr-2 cursor-pointer {tree[key].__visible ? 'rotate-180' : 'rotate-90'} transition" on:click={async () => onDirClick(`${path}/${key}`, tree[key].__fetched, tree[key].__visible)}>
          <Arrow />
        </div>
        <p class="flex-1 h-5 leading-4 {((included || paths.includes(`${path}/${key}`)) && 'text-green-500') || paths.find((p) => p.startsWith(`${path}/${key}`)) && 'text-blue-500' || ''}">
          <span class="cursor-pointer" on:click={() => onTogglePaths([`${path}/${key}`], true)}>{key}</span>
        </p>
      </div>
      <div class={`pl-5 ${tree[key].__visible ? '' : 'hidden'}`}>
        <AddPaths path={`${path}/${key}`} tree={tree[key]} {paths} {onTogglePaths} {onDirClick} />
      </div>
    {/if}
  {/each}
</div>

