<script lang="ts">
  import Archive from '@lucide/svelte/icons/archive';
  import ArchiveRestore from '@lucide/svelte/icons/archive-restore';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import LockOpen from '@lucide/svelte/icons/lock-open';
  import Scissors from '@lucide/svelte/icons/scissors';
  import ShieldCheck from '@lucide/svelte/icons/shield-check';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import { slide } from 'svelte/transition';

  import { api, type Job, type Plan } from '#lib/api/client.ts';
  import { jobs, notifyError } from '#lib/app.svelte.ts';
  import { duration, elapsed } from '#lib/format.ts';
  import { isActive } from '#lib/jobs.svelte.ts';
  import { KIND_LABELS, summarize } from '#lib/jobtext.ts';
  import { ms } from '#lib/motion.ts';
  import { byDay } from '#lib/timeline.ts';

  import JobBadge from './JobBadge.svelte';
  import JobProgress from './JobProgress.svelte';

  // Everything that ran on a repository, grouped by day, newest first.
  // plans name the plan of backups and retention runs when there are several
  let { repositoryId, plans = [] }: { repositoryId: number; plans?: Plan[] } = $props();
  const planName = (job: Job) => (plans.length > 1 ? plans.find((p) => p.id === job.plan_id)?.name : undefined);

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

  const days = $derived(byDay(jobs.list(repositoryId)));

  const ICONS = { backup: Archive, restore: ArchiveRestore, forget: Scissors, prune: Scissors, check: ShieldCheck, unlock: LockOpen, init: Sparkles };
  const DOT: Record<Job['status'], string> = {
    queued: 'bg-base-300 text-base-content',
    running: 'bg-info text-info-content',
    succeeded: 'bg-success text-success-content',
    failed: 'bg-error text-error-content',
    cancelled: 'bg-warning text-warning-content',
    interrupted: 'bg-warning text-warning-content',
  };
  const time = new Intl.DateTimeFormat(undefined, { timeStyle: 'short' });
</script>

<div data-testid="run-history">
  {#each days as day (day.key)}
    <section class="mb-6 last:mb-0">
      <h3 class="text-base-content/70 bg-base-100 sticky top-14 z-10 mb-2 py-1 text-xs font-semibold tracking-wide uppercase">{day.label}</h3>
      <ol class="border-base-300 ml-3 border-l">
        {#each day.jobs as job (job.id)}
          {@const Icon = ICONS[job.kind]}
          {@const open = expanded === job.id || isActive(job.status)}
          <li class="relative pb-3 pl-6 last:pb-0">
            <span class="ring-base-100 absolute top-1.5 -left-3 grid size-6 place-items-center rounded-full ring-4 {DOT[job.status]}" aria-hidden="true">
              <Icon class="size-3.5" />
            </span>
            <button
              class="hover:bg-base-200 flex w-full flex-wrap items-center gap-x-3 gap-y-1 rounded-lg px-2 py-1.5 text-left"
              aria-expanded={open}
              onclick={() => (expanded = expanded === job.id ? null : job.id)}
            >
              <span class="font-medium">{KIND_LABELS[job.kind]}</span>
              {#if planName(job)}<span class="badge badge-outline badge-sm" data-testid="job-plan">{planName(job)}</span>{/if}
              <JobBadge status={job.status} />
              {#if job.trigger === 'schedule'}<span class="badge badge-ghost badge-xs">scheduled</span>{/if}
              <span class="min-w-0 flex-1 truncate text-sm {job.error ? 'text-error' : 'text-base-content/80'}">{job.error || summarize(job)}</span>
              <span class="tabular text-base-content/70 text-xs whitespace-nowrap">
                <time datetime={job.started_at ?? job.created_at}>{time.format(Date.parse(job.started_at ?? job.created_at))}</time>
                {#if job.finished_at}· {duration(elapsed(job.started_at, job.finished_at))}{/if}
              </span>
              <ChevronRight class="text-base-content/50 size-4 transition-transform {open ? 'rotate-90' : ''}" aria-hidden="true" />
            </button>
            {#if open}
              <div class="mt-1 ml-2" transition:slide={{ duration: ms(180) }}>
                {#if isActive(job.status)}
                  <JobProgress {job} />
                {:else}
                  <div class="flex flex-col gap-1 text-xs">
                    {#if job.snapshot_id}<div>Snapshot <span class="font-mono">{job.snapshot_id.slice(0, 8)}</span></div>{/if}
                    {#if job.exit_code !== undefined}<div>restic exited with code {job.exit_code}</div>{/if}
                    {#if job.log_excerpt}
                      <pre class="bg-base-200 rounded-box max-h-64 overflow-auto p-2 whitespace-pre-wrap">{job.log_excerpt}</pre>
                    {:else}
                      <div class="text-base-content/70">No log output.</div>
                    {/if}
                  </div>
                {/if}
              </div>
            {/if}
          </li>
        {/each}
      </ol>
    </section>
  {:else}
    {#if loading}
      <div class="flex flex-col gap-3" aria-busy="true">
        {#each [1, 2, 3] as i (i)}<div class="skeleton h-10"></div>{/each}
      </div>
    {:else}
      <p class="text-base-content/70">Nothing has run yet. Backups, checks and restores show up here.</p>
    {/if}
  {/each}
  {#if more && days.length}
    <button class="btn btn-ghost btn-sm mt-4" onclick={load} disabled={loading}>Load older runs</button>
  {/if}
</div>
