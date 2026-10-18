<script lang="ts">
  import { resolve } from '$app/paths';
  import { api, type Repository } from '#lib/api/client.ts';
  import { jobs, notifyError } from '#lib/app.svelte.ts';
  import RepositoryCard from '#lib/components/RepositoryCard.svelte';

  let repositories = $state<Repository[] | null>(null);

  async function load() {
    try {
      const list = await api.repositories();
      // enough history for the last backup and anything running
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
    // a finished init or backup changes what the cards show
    const unsubscribe = jobs.onChange((job) => {
      if (job.status === 'succeeded' && job.kind === 'init') load();
    });
    return () => {
      stop();
      unsubscribe();
    };
  });
</script>

<svelte:head><title>Repositories · ResticGUI</title></svelte:head>

{#if repositories === null}
  <span class="loading loading-spinner"></span>
{:else if repositories.length === 0}
  <div class="hero bg-base-100 rounded-box py-16">
    <div class="hero-content text-center">
      <div class="max-w-md">
        <h1 class="text-2xl font-bold">No repository yet</h1>
        <p class="py-4">A repository is where restic keeps your backups: a folder, a bucket or a server.</p>
        <a class="btn btn-primary" href={resolve('/repositories/new')}>Add a repository</a>
      </div>
    </div>
  </div>
{:else}
  <div class="grid gap-4 md:grid-cols-2">
    {#each repositories as r (r.id)}
      <RepositoryCard repository={r} />
    {/each}
  </div>
{/if}
