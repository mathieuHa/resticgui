<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let page: number;
  export let itemPerPage = 20;
  export let numberItem = 0;


  const dispatch = createEventDispatcher();
</script>

<div class="flex justify-center gap-2">
  <button
    class="{page < 1 && 'hidden'} bg-blue-500 hover:bg-blue-600 transition px-1 inline rounded"
    on:click={() => dispatch('change', page - 1)}
  >
    previous 
  </button>
  <button
    class="{numberItem < itemPerPage && 'hidden'} bg-blue-500 hover:bg-blue-600 transition px-1 inline rounded"
    on:click={() => dispatch('change', page + 1)}
  >
    next
  </button>
</div>