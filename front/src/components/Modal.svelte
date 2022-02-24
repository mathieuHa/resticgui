<script>
  import { createEventDispatcher, tick } from 'svelte';
  import Cross from './Icon/Cross.svelte';

  function portal(node) {
    const anchor = document.getElementById('modals');
    if (anchor) anchor.insertBefore(node, null);
    return { destroy: () => anchor.contains(node) && anchor.removeChild(node) };
  }

  const dispatch = createEventDispatcher();
  const onClose = () => dispatch('close', true);
</script>

<div
  use:portal
  class="p-6 fixed top-0 left-0 right-0 bottom-0 bg-gray-200/[.5] dark:bg-gray-900/[.5]"
  on:click|self={onClose}
>
  <div class="max-h-full overflow-y-auto relative max-w-xl p-8 rounded-md m-auto bg-white dark:bg-gray-800 pointer-events-auto shadow-lg shadow-gray-300 dark:shadow-black">
    <span on:click={onClose} class="absolute top-2 right-2 cursor-pointer w-5 h-5">
      <Cross />
    </span>
    <slot />
  </div>
</div>
