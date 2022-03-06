<script>
  import { toasts } from '../utils/stores';

  $:console.log($toasts);
</script>

{#if $toasts.length}
  <div class="p-6 fixed top-0 right-0 pointer-events-none max-w-[300px] flex flex-col justify-end">
    {#each $toasts as toast}
      <div class="relative toaster ml-auto inline-flex mb-2 bg-white dark:bg-gray-800 pointer-events-auto rounded-md shadow-lg shadow-gray-400 dark:shadow-black">
        <span class="flex items-center justify-center min-h-full w-6 rounded-l bg-{(toast.type === 'success' && 'blue') || (toast.type === 'failure' && 'orange') || 'red'}-500">
          {(toast.type === 'success' && '🙂') || (toast.type === 'failure' && '☹') || '🤔'}
        </span>
        <p class="flex-1 px-2 py-1">{toast.value}</p>
      </div>
    {/each}
  </div>
{/if}

<style>
  @keyframes translation {
    from { right: -100%; opacity: 0; }
    to { right: 0; opacity: 1; }
  }
  .toaster {
    animation: translation 250ms;
  }
</style>
