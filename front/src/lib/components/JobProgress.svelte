<script lang="ts">
  import { api, type Job } from '#lib/api/client.ts';
  import { attempt, jobs } from '#lib/app.svelte.ts';
  import { bytes, duration, elapsed, percent } from '#lib/format.ts';
  import { KIND_LABELS } from '#lib/jobtext.ts';

  import JobBadge from './JobBadge.svelte';

  let { job }: { job: Job } = $props();

  const progress = $derived(jobs.progress[job.id]);
  let cancelling = $state(false);
  let now = $state(Date.now());

  $effect(() => {
    const timer = setInterval(() => (now = Date.now()), 1000);
    return () => clearInterval(timer);
  });

  async function cancel() {
    cancelling = true;
    await attempt(() => api.cancelJob(job.id), 'Could not cancel');
  }
</script>

<div class="flex flex-col gap-2" data-testid="job-progress">
  <div class="flex flex-wrap items-center gap-2 text-sm">
    <span class="font-medium">{KIND_LABELS[job.kind]}</span>
    <JobBadge status={job.status} />
    {#if job.status === 'running'}
      <span class="opacity-70">for {duration(elapsed(job.started_at, null, now))}</span>
    {:else}
      <span class="opacity-70">waiting for the repository</span>
    {/if}
    <button class="btn btn-ghost btn-xs ml-auto" onclick={cancel} disabled={cancelling}>
      {cancelling ? 'Cancelling…' : 'Cancel'}
    </button>
  </div>
  {#if progress}
    <progress class="progress progress-info w-full" value={progress.percent_done} max="1"></progress>
    <div class="flex flex-wrap gap-x-4 text-xs opacity-80">
      <span>{percent(progress.percent_done)}</span>
      <span>{progress.files_done} / {progress.total_files} files</span>
      <span>{bytes(progress.bytes_done)} / {bytes(progress.total_bytes)}</span>
      {#if progress.seconds_remaining}<span>{duration(progress.seconds_remaining)} left</span>{/if}
    </div>
    {#if progress.current_files?.length}
      <div class="truncate font-mono text-xs opacity-60" title={progress.current_files[0]}>{progress.current_files[0]}</div>
    {/if}
  {:else if job.status === 'running'}
    <progress class="progress w-full"></progress>
  {/if}
</div>
