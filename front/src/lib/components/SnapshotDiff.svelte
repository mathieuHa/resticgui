<script lang="ts">
  import { api, type Diff, type Snapshot } from '#lib/api/client.ts';
  import { changes, counts, type ChangeKind } from '#lib/diff.ts';
  import { bytes, dateTime } from '#lib/format.ts';

  let { repositoryId, from, to }: { repositoryId: number; from: Snapshot; to: Snapshot } = $props();

  let diff = $state<Diff | null>(null);
  let error = $state('');
  // long lists are cut, the counts above them stay whole
  const SHOWN = 500;

  $effect(() => {
    diff = null;
    error = '';
    api.diff(repositoryId, from.id, to.id).then(
      (d) => (diff = d),
      (err) => (error = err instanceof Error ? err.message : String(err)),
    );
  });

  const files = (n: number) => `${n} ${n === 1 ? 'file' : 'files'}`;
  const list = $derived(diff ? changes(diff.changes) : []);
  const n = $derived(counts(list));
  const BADGES: Record<ChangeKind, { label: string; tone: string }> = {
    added: { label: 'added', tone: 'badge-success' },
    removed: { label: 'removed', tone: 'badge-error' },
    changed: { label: 'changed', tone: 'badge-info' },
  };
</script>

<p class="text-base-content/70 text-sm">
  From <span class="font-mono">{from.short_id}</span> ({dateTime(from.time)}) to <span class="font-mono">{to.short_id}</span> ({dateTime(to.time)}).
</p>
{#if error}
  <div class="alert alert-error text-sm">{error}</div>
{:else if !diff}
  <div class="flex flex-col gap-2" aria-busy="true">{#each [1, 2, 3] as i (i)}<div class="skeleton h-6"></div>{/each}</div>
{:else}
  <div class="grid grid-cols-1 gap-2 sm:grid-cols-3" data-testid="diff-stats">
    <div class="bg-base-200 rounded-box p-3">
      <div class="text-base-content/70 text-xs">Added</div>
      <div class="tabular font-semibold">{files(diff.stats.added.files)}, {bytes(diff.stats.added.bytes)}</div>
    </div>
    <div class="bg-base-200 rounded-box p-3">
      <div class="text-base-content/70 text-xs">Removed</div>
      <div class="tabular font-semibold">{files(diff.stats.removed.files)}, {bytes(diff.stats.removed.bytes)}</div>
    </div>
    <div class="bg-base-200 rounded-box p-3">
      <div class="text-base-content/70 text-xs">Changed</div>
      <div class="tabular font-semibold">{files(diff.stats.changed_files)}</div>
    </div>
  </div>
  {#if list.length === 0}
    <p class="text-base-content/70 mt-2">Nothing changed between these snapshots.</p>
  {:else}
    <p class="mt-2 text-sm">{n.added} added, {n.removed} removed, {n.changed} changed.</p>
    <ul class="mt-1 flex max-h-96 flex-col gap-1 overflow-y-auto" data-testid="diff-changes">
      {#each list.slice(0, SHOWN) as c (c.path)}
        <li class="flex min-w-0 items-center gap-2 text-sm">
          <span class="badge badge-sm {BADGES[c.kind].tone} badge-soft w-18 shrink-0">{BADGES[c.kind].label}</span>
          <span class="truncate font-mono text-xs" title={c.path}>{c.path}{c.dir ? '/' : ''}</span>
        </li>
      {/each}
    </ul>
    {#if list.length > SHOWN}<p class="text-base-content/70 text-xs">And {list.length - SHOWN} more.</p>{/if}
  {/if}
{/if}
