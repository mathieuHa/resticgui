<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let page: number;
  export let itemPerPage = 20;
  export let numberItem = 0;


  const dispatch = createEventDispatcher();
</script>

<div class="flex justify-center gap-2 mt-4">
  <button
    class="btn btn-ghost {page < 1 && 'hidden'}"
    on:click={() => dispatch('change', page - 1)}
  >
    previous 
  </button>
  <button
    class="btn btn-ghost {numberItem < itemPerPage && 'hidden'}"
    on:click={() => dispatch('change', page + 1)}
  >
    next
  </button>
</div>