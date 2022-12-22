<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import agent from '$utils/api';
  import AddPaths from './AddPaths.svelte';
  import type { Path } from '$utils/types';
  
  interface Tree {
    [key: string]: Tree | boolean
  }

  export let initialValues: Path[] = [];
  let tree: Tree = {};
  let paths: string[] = initialValues?.map(
    ({ relPath = '/root', included }) => included ? relPath : `!${relPath}`,
  );

  const leaves = {};
  const dispatch = createEventDispatcher();

  const populateTree = (treeCopy?: Tree, root?: string[], leaf?: Tree | boolean): Tree | undefined => {
    if (!treeCopy || !root || leaf === undefined) throw new Error()
    if (root.length === 0) return leaf;
    const path = root.shift() ?? '';
    if (root.length === 0) return { ...treeCopy, [path]: leaf };
    return { ...treeCopy, [path]: { __visible: true, __fetched: true, ...populateTree(treeCopy[path], root, leaf) } };
  }

  const onLoad = async (root: string): Promise<void> => {
    const rootArray = root.split('/').filter((path) => !!path);
    let result;
    try {
      result = await agent.post('/paths', { name: root }) || [];
    } catch (err) {
      result = [];
    }
    const leaf = { __fetched: true, __visible: true };
    for (let value of result) {
      leaves[`${root}${value.name}`] = value;
      if (value.isDir) leaf[value.name] = { __fetched: false, __visible: false };
      else leaf[value.name] = value.name;
    }
    tree = populateTree(tree, rootArray, leaf);
  }

  const onTogglePaths = (path: string, include = true) => {
    const index = paths.findIndex((p) => p === path || p === `!${path}`);
    if (index > -1) paths = [...paths.slice(0, index), ...paths.slice(index + 1)];
    else paths = [...paths, `${include ? '' : '!'}${path}`];

    for (let i = paths.length - 1; i >= 0; i -= 1) {
      if (paths[i] !== path && paths[i] !== `!${path}`
        && ((paths[i][0] === '!' && `${paths[i]}/`.slice(1).startsWith(`${path}/`))
        || (`${paths[i]}/`.startsWith(`${path}/`)))) {
        paths = [...paths.slice(0, i), ...paths.slice(i + 1)];
      }
    }
  };

  const onDirClick = async (root: string, fetched: boolean, visible: boolean) => {
    const rootArray = root.split('/').filter((path) => !!path);
    if (fetched) tree = populateTree(tree, [...rootArray, '__visible'], !visible);
    else await onLoad(root);
  };

  onMount(() => onLoad('/'));
  $:dispatch('update', paths.map((relPath) => relPath[0] === '!'
    ? {
      ...leaves[relPath.slice(1)] ?? initialValues.find((val) => val.relPath === relPath.slice(1)),
      included: false,
      relPath: relPath.slice(1),
    } : {
      ...leaves[relPath] ?? initialValues.find((val) => val.relPath === relPath),
      included: true,
      relPath,
    }));
</script>

<p class="font-bold">Paths to backup</p>
<p class="py-2 bg-white dark:bg-gray-800 border border-x-0 border-t-0 border-stone-500">
  [
  {#each paths as path, i}
    "<span class={path[0] === '!' ? 'text-orange-500' : ''}>{path[0] === '!' ? path.slice(1) : path}</span>"{i !== paths.length - 1 ? ', ' : ''}
  {/each}
  ]
</p>
<div class="overflow-y-auto">
  <AddPaths
    path=""
    tree={tree}
    {paths}
    {onTogglePaths}
    {onDirClick}
  />
</div>
