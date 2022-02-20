<script>
  import { createEventDispatcher, onMount } from 'svelte';

  import agent from '../utils/api';
  import AddPaths from './AddPaths.svelte';

  export let initialPaths;
  let tree = {};
  let paths = initialPaths || [];

  const dispatch = createEventDispatcher();

  const populateTree = (treeCopy, root, leaf) => {
    if (root.length === 0) return leaf;
    const path = root.shift();
    if (root.length === 0) return { ...treeCopy, [path]: leaf };
    return { ...treeCopy, [path]: { __visible: true, __fetched: true, ...populateTree(treeCopy[path], root, leaf) } };
  }

  const onLoad = async (root) => {
    const rootArray = root.split('/').filter((path) => !!path);
    let result;
    try {
      result = await agent.post('/paths', { name: root }) || [];
    } catch (err) {
      result = [];
    }
    const leaf = { __fetched: true, __visible: true };
    for (let value of result) {
      if (value.isDir) leaf[value.name] = { __fetched: false, __visible: false };
      else leaf[value.name] = value.name;
    }
    tree = populateTree(tree, rootArray, leaf);
  }

  const onTogglePaths = (pathArray, isDir) => {
    for (let path of pathArray) {
      const index = paths.findIndex((p) => p === path);
      if (index > -1) paths = [...paths.slice(0, index), ...paths.slice(index + 1)];
      else paths = [...paths, path];
    }
  };

  const onDirClick = (root, fetched, visible) => {
    const rootArray = root.split('/').filter((path) => !!path);
    if (fetched) tree = populateTree(tree, [...rootArray, '__visible'], !visible)
    else onLoad(root)
  };

  onMount(() => onLoad('/'));
  $:dispatch('update', paths);
  $:console.log(paths, tree);
</script>

<AddPaths path="" tree={tree} {paths} {onTogglePaths} {onDirClick} />
