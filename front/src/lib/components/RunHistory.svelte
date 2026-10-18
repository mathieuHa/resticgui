<script lang="ts">
  import { api } from '#lib/api/client.ts';
  import { jobs, notifyError } from '#lib/app.svelte.ts';
  import { dateTime, duration, elapsed } from '#lib/format.ts';
  import { KIND_LABELS, summarize } from '#lib/jobtext.ts';

  import JobBadge from './JobBadge.svelte';
  import JobProgress from './JobProgress.svelte';

  // The jobs of a repository, newest first, with their outcome and log.
  let { repositoryId }: { repositoryId: number } = $props();

  const PAGE = 25;
  let loaded = $state(0);
  let more = $state(true);
  let loading = $state(false);
  let expanded = $state<number | null>(null);

  async function load() {
    loading = true;
    try {
      const page = await api.jobs(repositoryId, PAGE, loaded);
      jobs.add(page);
      loaded += page.length;
      more = page.length === PAGE;
    } catch (err) {
      notifyError(err, 'Could not load the runs');
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    load();
  });

  const list = $derived(jobs.list(repositoryId));
</script>

<div class="overflow-x-auto" data-testid="run-history">
  <table class="table-sm table">
    <thead>
      <tr><th>Job</th><th>Status</th><th>Started</th><th>Took</th><th>Result</th></tr>
    </thead>
    <tbody>
      {#each list as job (job.id)}
        <tr class="hover:bg-base-200 cursor-pointer" onclick={() => (expanded = expanded === job.id ? null : job.id)}>
          <td class="whitespace-nowrap">
            {KIND_LABELS[job.kind]}
            {#if job.trigger === 'schedule'}<span class="badge badge-ghost badge-xs">scheduled</span>{/if}
          </td>
          <td><JobBadge status={job.status} /></td>
          <td class="whitespace-nowrap">{dateTime(job.started_at ?? job.created_at)}</td>
          <td class="whitespace-nowrap">{job.finished_at ? duration(elapsed(job.started_at, job.finished_at)) : ''}</td>
          <td class="max-w-md truncate {job.error ? 'text-error' : ''}">{job.error || summarize(job)}</td>
        </tr>
        {#if expanded === job.id || job.status === 'running' || job.status === 'queued'}
          <tr>
            <td colspan="5" class="bg-base-200/50">
              {#if job.status === 'running' || job.status === 'queued'}
                <JobProgress {job} />
              {:else}
                <div class="flex flex-col gap-1 text-xs">
                  {#if job.snapshot_id}<div>Snapshot <span class="font-mono">{job.snapshot_id.slice(0, 8)}</span></div>{/if}
                  {#if job.exit_code !== undefined}<div>restic exited with code {job.exit_code}</div>{/if}
                  {#if job.log_excerpt}
                    <pre class="bg-base-300 rounded-box max-h-64 overflow-auto p-2 whitespace-pre-wrap">{job.log_excerpt}</pre>
                  {:else}
                    <div class="opacity-60">No log output.</div>
                  {/if}
                </div>
              {/if}
            </td>
          </tr>
        {/if}
      {:else}
        <tr><td colspan="5" class="opacity-60">{loading ? 'Loading…' : 'Nothing has run yet.'}</td></tr>
      {/each}
    </tbody>
  </table>
  {#if more && list.length}
    <button class="btn btn-ghost btn-sm mt-2" onclick={load} disabled={loading}>Load older runs</button>
  {/if}
</div>
