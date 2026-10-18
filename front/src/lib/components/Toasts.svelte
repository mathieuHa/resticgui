<script lang="ts">
  import { dismiss, toasts } from '$lib/app.svelte';

  const classes = { success: 'alert-success', error: 'alert-error', info: 'alert-info' };
</script>

<div class="toast toast-end z-50">
  {#each toasts as t (t.id)}
    <div role="alert" class="alert {classes[t.kind]} max-w-md">
      <div class="flex flex-col">
        {#each t.lines as line, i (i)}
          <span class={i ? 'text-sm' : 'font-medium'}>{line}</span>
        {/each}
      </div>
      <button class="btn btn-ghost btn-xs" aria-label="Dismiss" onclick={() => dismiss(t.id)}>✕</button>
    </div>
  {/each}
</div>
