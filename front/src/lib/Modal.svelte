<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Cross from './Icon/Cross.svelte';

  export let title: string;

  function portal(node: HTMLDivElement) {
    const anchor = document.getElementById('modals');
    if (anchor) anchor.insertBefore(node, null);
    return { destroy: () => anchor?.contains(node) && anchor.removeChild(node) };
  }

  const dispatch = createEventDispatcher();
  const onClose = () => dispatch('close', true);
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div
  use:portal
  class="p-6 fixed top-0 left-0 right-0 bottom-0 bg-gray-200/[.5] dark:bg-gray-900/[.5]"
  on:click|self={onClose}
>
  <div class="custom-max-height overflow-y-auto relative max-w-xl bg-gray-200 dark:bg-gray-700 rounded-md m-auto shadow-lg shadow-gray-400 dark:shadow-black">
    <div class="flex bg-white dark:bg-gray-800">
      <p class="flex-1 px-2 py-1">{title}</p>
      <button on:click|preventDefault|stopPropagation={onClose} class="hover:bg-red-600 p-2 cursor-pointer transition-all">
        <Cross />
      </button>
    </div>
    <div class="p-2">
      <slot />
    </div>
  </div>
</div>

<style>
  .custom-max-height {
    max-width: 700px;
  }
</style>
