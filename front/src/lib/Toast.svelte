<script lang="ts">
  import { toasts } from '$utils/stores';
</script>

{#if $toasts.length}
  <div class="p-6 fixed top-0 right-0 pointer-events-none max-w-[300px] flex flex-col justify-end">
    {#each $toasts as toast}
      <div id="toast_{toast.id}" class="relative toaster ml-auto inline-flex mb-2 bg-base-100 pointer-events-auto rounded-xl shadow-base-100">
        <span class="flex items-center justify-center min-h-full w-8 rounded-l-xl pl-1 bg-{(toast.type === 'success' && 'blue') || (toast.type === 'failure' && 'orange') || 'red'}-500">
          {(toast.type === 'success' && '🙂') || (toast.type === 'failure' && '☹') || '🤔'}
        </span>
        <p class="flex-1 px-2 py-1">{toast.value}</p>
      </div>
    {/each}
  </div>
{/if}

<style>
  @keyframes translation {
    0% { right: -100%; opacity: 0; }
    10% { right: 0; opacity: 1; }
    90% { right: 0; opacity: 1; }
    100% { right: -100%; opacity: 0; }
  }
  .toaster {
    animation: translation 5000ms;
  }
</style>
