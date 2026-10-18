<script lang="ts">
  import type { BackupPoint } from '#lib/charts.ts';
  import { dateTime } from '#lib/format.ts';

  // The outcome of the last backups as a row of ticks, oldest on the left.
  let { points, slots = 20 }: { points: BackupPoint[]; slots?: number } = $props();

  const COLOR: Partial<Record<BackupPoint['status'], string>> = {
    succeeded: 'bg-success',
    failed: 'bg-error',
    interrupted: 'bg-error',
    cancelled: 'bg-warning',
  };

  const shown = $derived(points.slice(-slots));
  const empty = $derived(Math.max(0, slots - shown.length));
  const good = $derived(shown.filter((p) => p.status === 'succeeded').length);
</script>

<div class="flex h-5 items-end gap-0.5" role="img" aria-label="{good} of the last {shown.length} backups succeeded" data-testid="status-strip">
  {#each { length: empty }, i (i)}
    <span class="bg-base-300 h-2 flex-1 rounded-sm"></span>
  {/each}
  {#each shown as p (p.id)}
    <span class="h-full flex-1 rounded-sm {COLOR[p.status] ?? 'bg-base-300'}" title="{dateTime(new Date(p.time).toISOString())}: {p.status}"></span>
  {/each}
</div>
