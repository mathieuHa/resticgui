<script lang="ts">
  import Ban from '@lucide/svelte/icons/ban';
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import CirclePause from '@lucide/svelte/icons/circle-pause';
  import Hourglass from '@lucide/svelte/icons/hourglass';
  import Loader from '@lucide/svelte/icons/loader';

  import type { JobStatus } from '#lib/api/client.ts';

  let { status }: { status: JobStatus } = $props();

  const LOOK = {
    queued: { cls: 'badge-ghost', icon: Hourglass },
    running: { cls: 'badge-info', icon: Loader },
    succeeded: { cls: 'badge-success', icon: CircleCheck },
    failed: { cls: 'badge-error', icon: CircleAlert },
    cancelled: { cls: 'badge-warning', icon: Ban },
    interrupted: { cls: 'badge-warning', icon: CirclePause },
  } as const;
  const look = $derived(LOOK[status]);
</script>

<span class="badge badge-sm {look.cls} gap-1" data-status={status}>
  <look.icon class="size-3 {status === 'running' ? 'animate-spin' : ''}" aria-hidden="true" />
  {status}
</span>
