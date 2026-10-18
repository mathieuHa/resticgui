<script lang="ts">
  import type { JobStatus } from '#lib/api/client.ts';

  let { status }: { status: JobStatus } = $props();

  const classes: Record<JobStatus, string> = {
    queued: 'badge-ghost',
    running: 'badge-info',
    succeeded: 'badge-success',
    failed: 'badge-error',
    cancelled: 'badge-warning',
    interrupted: 'badge-warning',
  };
</script>

<span class="badge badge-sm {classes[status]}" data-status={status}>
  {#if status === 'running'}<span class="loading loading-spinner loading-xs"></span>{/if}
  {status}
</span>
