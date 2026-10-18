<script lang="ts">
  import Check from '@lucide/svelte/icons/check';
  import type { Snippet } from 'svelte';

  // Numbered steps; the first one not done is the current one.
  let { steps }: { steps: { title: string; done: boolean; action?: Snippet }[] } = $props();

  const current = $derived(steps.findIndex((s) => !s.done));
</script>

<ol class="flex flex-col gap-2" data-testid="checklist">
  {#each steps as step, i (step.title)}
    <li class="flex items-center gap-3 {i === current ? '' : 'text-base-content/70'}" aria-current={i === current ? 'step' : undefined}>
      <span
        class="grid size-6 shrink-0 place-items-center rounded-full text-xs font-semibold
          {step.done ? 'bg-success text-success-content' : i === current ? 'bg-primary text-primary-content' : 'bg-base-300'}"
      >
        {#if step.done}<Check class="size-3.5" aria-label="done" />{:else}{i + 1}{/if}
      </span>
      <span class="flex-1 {i === current ? 'font-medium' : ''}">{step.title}</span>
      {#if i === current && step.action}{@render step.action()}{/if}
    </li>
  {/each}
</ol>
