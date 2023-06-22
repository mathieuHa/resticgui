<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<dialog
  class="modal modal-open modal-top"
  on:click|stopPropagation={() => dispatch('close')}
>
  <div
    class="mx-auto w-11/12 max-w-2xl sm:px-8"
    on:click|stopPropagation={() => {}}
  >
    <slot />
  </div>
  
</dialog>