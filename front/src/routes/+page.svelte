<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import Info from '@lucide/svelte/icons/info';
  import Plus from '@lucide/svelte/icons/plus';
  import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
  import { resolve } from '$app/paths';
  import { flip } from 'svelte/animate';

  import { api, type Repository } from '#lib/api/client.ts';
  import { jobs, notifyError } from '#lib/app.svelte.ts';
  import Checklist from '#lib/components/Checklist.svelte';
  import Logo from '#lib/components/Logo.svelte';
  import RepositoryCard from '#lib/components/RepositoryCard.svelte';
  import { DASHBOARD_ORDER, headline, health } from '#lib/health.ts';
  import { ms } from '#lib/motion.ts';

  let repositories = $state<Repository[] | null>(null);
  let now = $state(Date.now());

  async function load() {
    try {
      const list = await api.repositories();
      // enough history for the status strip and anything running
      await Promise.all(list.map(async (r) => jobs.add(await api.jobs(r.id, 20))));
      repositories = list;
    } catch (err) {
      notifyError(err, 'Could not load the repositories');
      repositories ??= [];
    }
  }

  $effect(() => {
    load();
    const stop = jobs.onResync(load);
    // a finished job changes what the cards show, such as the next run
    const unsubscribe = jobs.onChange((job) => {
      if (job.status === 'succeeded' && (job.kind === 'init' || job.kind === 'backup')) load();
    });
    // relative times stay current
    const tick = setInterval(() => (now = Date.now()), 30_000);
    return () => {
      stop();
      unsubscribe();
      clearInterval(tick);
    };
  });

  const states = $derived(repositories?.map((r) => ({ r, h: health(r, jobs.list(r.id), now) })) ?? []);
  // the ones that need something first, then by name
  const sorted = $derived(
    [...states].sort((a, b) => DASHBOARD_ORDER.indexOf(a.h) - DASHBOARD_ORDER.indexOf(b.h) || a.r.name.localeCompare(b.r.name)),
  );
  const summary = $derived(headline(states.map((s) => s.h)));
  const TONE = {
    success: { cls: 'bg-success/10 text-success border-success/30', icon: CircleCheck },
    warning: { cls: 'bg-warning/15 border-warning/40', icon: TriangleAlert },
    error: { cls: 'bg-error/10 text-error border-error/30', icon: CircleAlert },
    info: { cls: 'bg-info/10 text-info border-info/30', icon: Info },
  };
  const tone = $derived(TONE[summary.tone]);
</script>

<svelte:head><title>Repositories · ResticGUI</title></svelte:head>

{#if repositories === null}
  <div class="skeleton mb-6 h-14 w-full"></div>
  <div class="grid grid-cols-1 gap-4 md:grid-cols-2" aria-busy="true" aria-label="Loading the repositories">
    {#each [1, 2] as i (i)}
      <div class="skeleton h-56"></div>
    {/each}
  </div>
{:else if repositories.length === 0}
  <section class="bg-base-100 border-base-300 animate-rise mx-auto mt-6 max-w-xl rounded-box border p-8 shadow-sm">
    <div class="mb-6 flex items-center gap-3">
      <Logo />
      <div>
        <h1 class="text-xl font-semibold">No repository yet</h1>
        <p class="text-base-content/70 text-sm">A repository is where restic keeps your backups: a folder, a bucket or a server.</p>
      </div>
    </div>
    <Checklist
      steps={[
        { title: 'Add a repository', done: false, action: add },
        { title: 'Choose what to back up', done: false },
        { title: 'Run the first backup', done: false },
      ]}
    />
    <p class="text-base-content/70 mt-6 text-sm">
      Backups then run on their schedule. This page tells you when one fails or is late.
    </p>
  </section>
{:else}
  <div class="border-l-4 {tone.cls} animate-rise mb-6 flex items-center gap-3 rounded-box border p-4" role="status" data-testid="headline">
    <tone.icon class="size-6 shrink-0" aria-hidden="true" />
    <p class="text-base-content font-medium">{summary.text}</p>
  </div>
  <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
    {#each sorted as { r } (r.id)}
      <div animate:flip={{ duration: ms(300) }}>
        <RepositoryCard repository={r} {now} />
      </div>
    {/each}
  </div>
{/if}

{#snippet add()}
  <a class="btn btn-primary btn-sm" href={resolve('/repositories/new')}><Plus class="size-4" aria-hidden="true" />Add a repository</a>
{/snippet}
