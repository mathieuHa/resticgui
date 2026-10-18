<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import CircleDashed from '@lucide/svelte/icons/circle-dashed';
  import Clock from '@lucide/svelte/icons/clock';
  import Loader from '@lucide/svelte/icons/loader';
  import Wrench from '@lucide/svelte/icons/wrench';

  import type { Health } from '#lib/health.ts';

  let { health }: { health: Health } = $props();

  const LOOK = {
    ok: { label: 'Backed up', icon: CircleCheck, cls: 'badge-success' },
    running: { label: 'Running', icon: Loader, cls: 'badge-info' },
    late: { label: 'Late', icon: Clock, cls: 'badge-warning' },
    failed: { label: 'Failed', icon: CircleAlert, cls: 'badge-error' },
    never: { label: 'No backup yet', icon: CircleDashed, cls: 'badge-warning' },
    setup: { label: 'Set up', icon: Wrench, cls: 'badge-neutral' },
  } as const;

  const look = $derived(LOOK[health]);
</script>

<span class="badge badge-sm {look.cls} gap-1 whitespace-nowrap" data-health={health}>
  <look.icon class="size-3.5 {health === 'running' ? 'animate-spin' : ''}" aria-hidden="true" />
  {look.label}
</span>
