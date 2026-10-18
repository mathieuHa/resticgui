<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import Info from '@lucide/svelte/icons/info';
  import X from '@lucide/svelte/icons/x';
  import { flip } from 'svelte/animate';
  import { fade, fly } from 'svelte/transition';

  import { dismiss, toasts } from '#lib/app.svelte.ts';
  import { ms } from '#lib/motion.ts';

  const classes = { success: 'alert-success', error: 'alert-error', info: 'alert-info' };
  const icons = { success: CircleCheck, error: CircleAlert, info: Info };
</script>

<div class="toast toast-end z-50">
  {#each toasts as t (t.id)}
    {@const Icon = icons[t.kind]}
    <div
      role={t.kind === 'error' ? 'alert' : 'status'}
      class="alert {classes[t.kind]} max-w-md shadow-lg"
      in:fly={{ x: 24, duration: ms(220) }}
      out:fade={{ duration: ms(150) }}
      animate:flip={{ duration: ms(200) }}
    >
      <Icon class="size-5 shrink-0" aria-hidden="true" />
      <div class="flex flex-col">
        {#each t.lines as line, i (i)}
          <span class={i ? 'text-sm' : 'font-medium'}>{line}</span>
        {/each}
      </div>
      <button class="btn btn-ghost btn-xs btn-square" aria-label="Dismiss" onclick={() => dismiss(t.id)}>
        <X class="size-4" aria-hidden="true" />
      </button>
    </div>
  {/each}
</div>
