<script lang="ts">
  import Monitor from '@lucide/svelte/icons/monitor';
  import Moon from '@lucide/svelte/icons/moon';
  import Sun from '@lucide/svelte/icons/sun';

  import { theme, type ThemeChoice } from '#lib/theme.svelte.ts';

  const CHOICES = [
    { value: 'system', label: 'System theme', icon: Monitor },
    { value: 'light', label: 'Light theme', icon: Sun },
    { value: 'dark', label: 'Dark theme', icon: Moon },
  ] as const satisfies { value: ThemeChoice; label: string; icon: unknown }[];
</script>

<div class="join" role="radiogroup" aria-label="Theme">
  {#each CHOICES as c (c.value)}
    <button
      class="btn btn-ghost btn-sm btn-square join-item {theme.choice === c.value ? 'btn-active' : ''}"
      role="radio"
      aria-checked={theme.choice === c.value}
      aria-label={c.label}
      title={c.label}
      onclick={() => theme.set(c.value)}
    >
      <c.icon class="size-4" aria-hidden="true" />
    </button>
  {/each}
</div>
