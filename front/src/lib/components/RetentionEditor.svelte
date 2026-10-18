<script lang="ts">
  import { api, type Retention, type RetentionPreview } from '#lib/api/client.ts';
  import { notifyError } from '#lib/app.svelte.ts';
  import { dateTime } from '#lib/format.ts';
  import { COUNT_RULES, isEmpty, normalize } from '#lib/retention.ts';

  // Edits a retention policy; with a repository id it can preview the policy.
  let {
    value = $bindable(),
    repositoryId,
  }: {
    value: Retention;
    repositoryId?: number;
  } = $props();

  let preview: RetentionPreview | null = $state(null);
  let previewing = $state(false);

  async function runPreview() {
    if (!repositoryId) return;
    previewing = true;
    try {
      preview = await api.previewRetention(repositoryId, normalize($state.snapshot(value)));
    } catch (err) {
      notifyError(err, 'Could not preview the policy');
    } finally {
      previewing = false;
    }
  }
</script>

<div class="flex flex-col gap-3">
  <p class="text-sm opacity-70">Keep the most recent snapshot of each period, for this many periods.</p>
  <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
    {#each COUNT_RULES as [key, label] (key)}
      <label class="input input-sm">
        <span class="label w-16">{label}</span>
        <input type="number" min="0" placeholder="0" bind:value={value[key]} />
      </label>
    {/each}
    <label class="input input-sm col-span-2 sm:col-span-3">
      <span class="label">Keep everything newer than</span>
      <input placeholder="e.g. 30d or 1y6m" bind:value={value.keep_within} />
    </label>
  </div>
  <label class="label cursor-pointer gap-2">
    <input type="checkbox" class="checkbox checkbox-sm" bind:checked={value.prune} />
    <span class="text-base-content">Prune afterwards, to free the space of removed snapshots</span>
  </label>
  {#if isEmpty(value)}
    <p class="text-warning text-sm">Set at least one rule.</p>
  {/if}

  {#if repositoryId}
    <div>
      <button type="button" class="btn btn-sm" onclick={runPreview} disabled={previewing || isEmpty(value)}>
        {previewing ? 'Checking snapshots…' : 'Preview'}
      </button>
    </div>
    {#if preview}
      <div class="grid gap-3 sm:grid-cols-2" data-testid="retention-preview">
        <div>
          <h4 class="text-success font-medium">Keeps {preview.keep.length}</h4>
          <ul class="text-sm">
            {#each preview.keep as s (s.id)}
              <li><span class="font-mono">{s.short_id}</span> {dateTime(s.time)} <span class="opacity-60">{s.matches?.join(', ')}</span></li>
            {/each}
          </ul>
        </div>
        <div>
          <h4 class="text-error font-medium">Removes {preview.remove.length}</h4>
          <ul class="text-sm">
            {#each preview.remove as s (s.id)}
              <li><span class="font-mono">{s.short_id}</span> {dateTime(s.time)}</li>
            {/each}
          </ul>
        </div>
      </div>
    {/if}
  {/if}
</div>
