<script lang="ts">
  import type { PlanDraft } from '#lib/plan.ts';

  import PathsEditor from './PathsEditor.svelte';
  import RetentionEditor from './RetentionEditor.svelte';
  import ScheduleEditor from './ScheduleEditor.svelte';

  // A backup plan: what to back up and how ('what'), when and for how long
  // ('when'), or both. For a saved plan, the retention can be previewed.
  let {
    draft = $bindable(),
    show = 'all',
    repositoryId,
    planId,
  }: {
    draft: PlanDraft;
    show?: 'what' | 'when' | 'all';
    repositoryId?: number;
    planId?: number;
  } = $props();

  let advanced = $state(false);
  const uid = $props.id();
</script>

<div class="flex flex-col gap-4">
  {#if show !== 'when'}
    <label class="floating-label">
      <span>Plan name</span>
      <input class="input w-full" placeholder="Plan name, such as Photos" bind:value={draft.name} data-testid="plan-name" />
    </label>
    <PathsEditor bind:value={draft.paths} />

    <div class="collapse-arrow bg-base-200 collapse {advanced ? 'collapse-open' : ''}">
      <button type="button" class="collapse-title text-left text-sm font-medium" aria-expanded={advanced} onclick={() => (advanced = !advanced)}>
        Backup options
      </button>
      <div class="collapse-content flex flex-col gap-3">
        <label class="input input-sm w-full">
          <span class="label">Tags</span>
          <input placeholder="e.g. laptop work" bind:value={draft.tags} />
        </label>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="select select-sm w-full">
            <span class="label">Compression</span>
            <select bind:value={draft.compression}>
              <option value="auto">auto</option>
              <option value="max">max</option>
              <option value="off">off</option>
            </select>
          </label>
          <label class="input input-sm w-full">
            <span class="label">Skip files larger than</span>
            <input placeholder="e.g. 2G" bind:value={draft.exclude_larger_than} />
          </label>
          <label class="input input-sm w-full">
            <span class="label">Upload limit</span>
            <input type="number" min="0" placeholder="none" bind:value={draft.limit_upload} />
            <span class="label">KiB/s</span>
          </label>
          <label class="input input-sm w-full">
            <span class="label">Download limit</span>
            <input type="number" min="0" placeholder="none" bind:value={draft.limit_download} />
            <span class="label">KiB/s</span>
          </label>
        </div>
        <label class="label cursor-pointer gap-2">
          <input type="checkbox" class="checkbox checkbox-sm" bind:checked={draft.one_file_system} />
          <span class="text-base-content">Stay on the file system of each folder</span>
        </label>
        <label class="label cursor-pointer gap-2">
          <input type="checkbox" class="checkbox checkbox-sm" bind:checked={draft.exclude_caches} />
          <span class="text-base-content">Skip cache folders (marked with a CACHEDIR.TAG file)</span>
        </label>
        <label class="flex flex-col gap-1 text-sm" for="{uid}-present">
          <span>Skip folders that hold a file named, one per line</span>
          <textarea id="{uid}-present" class="textarea textarea-sm w-full font-mono" rows="2" placeholder=".nobackup" bind:value={draft.exclude_if_present}></textarea>
        </label>
      </div>
    </div>
  {/if}

  {#if show !== 'what'}
    <ScheduleEditor label="Back up automatically" bind:value={draft.backupCron} />
    <div class="divider my-0"></div>
    <label class="label cursor-pointer gap-3">
      <input type="checkbox" class="toggle toggle-sm" bind:checked={draft.retentionOn} />
      <span class="text-base-content font-medium">Remove old snapshots</span>
    </label>
    {#if draft.retentionOn}
      <RetentionEditor bind:value={draft.retention} {repositoryId} {planId} />
      <ScheduleEditor label="Apply the retention policy automatically" bind:value={draft.forgetCron} />
    {/if}
  {/if}
</div>
