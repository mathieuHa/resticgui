<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<dialog
  class="modal modal-open modal-top"
  on:click|preventDefault|stopPropagation={() => dispatch('close')}
>
  <div class="mx-auto w-11/12 max-w-2xl px-8">
    <slot />
  </div>
  
</dialog>