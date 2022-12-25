<script lang="ts">
  import { onDestroy } from 'svelte';

  import { page } from '$app/stores';

  import Pagination from '$lib/Pagination.svelte';
  import Log from '$lib/RepositoryId/Log.svelte';
  import Snapshot from '$lib/RepositoryId/Snapshots/Snapshot.svelte';
  import General from '$lib/RepositoryId/General.svelte';
  import Loader from '$lib/Loader.svelte';
  import agent from '$utils/api';
  import { refresh } from '$utils/stores';

  const values = {
    tab: $page.url.searchParams.get('tab') || 'general',
    ps: $page.url.searchParams.get('ps') || '0',
    pl: $page.url.searchParams.get('pl') || '0',
  }

  const fetch = async (path: string) => agent.get(path);
  let repoPromise = fetch(`/repositories/${$page.params.id}`);
  let logsPromise = fetch(`/repositories/${$page.params.id}/logs?page=${values.pl}`);
  let snapshotsPromise = fetch(`/repositories/${$page.params.id}/snapshots?page=${values.ps}`);
  
  onDestroy(refresh.subscribe(() => {
    repoPromise = fetch(`/repositories/${$page.params.id}`);
    logsPromise = fetch(`/repositories/${$page.params.id}/logs?page=${values.pl}`);
    snapshotsPromise = fetch(`/repositories/${$page.params.id}/snapshots?page=${values.ps}`);
  }));
  
  const replaceStateWithQuery = (key: keyof typeof values, value: string) => {
    const url = new URL(window.location.toString());
    url.searchParams.set(key, value);
    history.replaceState({}, '', url);
    values[key] = value;
    if (key == 'ps') {
      snapshotsPromise = fetch(`/repositories/${$page.params.id}/snapshots?page=${values.ps}`);
    }
    if (key == 'pl') {
      logsPromise = fetch(`/repositories/${$page.params.id}/logs?page=${values.pl}`);
    }
  };
</script>

<div class="mt-2 mb-10 w-full max-w-full">
  <div class="flex justify-between flex-wrap mb-5">
    {#each ["General", "Snapshots", "Logs"] as stepped, i}
      <button
        on:click={() => replaceStateWithQuery('tab', stepped.toLowerCase())}
        class="flex-1 pb-2 px-4 {values.tab === stepped.toLowerCase()
          ? 'border-b-2 border-gray-500 order-4 md:order-unset'
          : 'cursor-pointer hover:border-b-2 hover:border-b-blue-500/[.5]'}"
      >
        <span class="rounded-xl">{stepped}</span>
      </button>
    {/each}
  </div>

  {#if values.tab === 'snapshots'}
    {#await snapshotsPromise}
      <Loader />
    {:then snapshots}
      {#if snapshots && snapshots.length > 0}
        <Snapshot 
          on:delete={() => (snapshotsPromise = fetch(`/repositories/${$page.params.id}/snapshots?page=${values.pl}`))}
          snapshots={snapshots}
        />
        <Pagination
          numberItem={snapshots.length}
          page={Number(values.ps)}
          on:change={({ detail }) => replaceStateWithQuery('ps', detail)}
        />
      {:else}
        <span>no data</span>
      {/if}
    {:catch err}
      <pre>{err}</pre>
    {/await}
  {:else if values.tab === 'logs'}
    {#await logsPromise}
      <Loader />
    {:then logs}
      {#if logs && logs.length > 0}
        <Log {logs} />
        <Pagination
          itemPerPage={10}
          numberItem={logs.length}
          page={Number(values.pl)}
          on:change={({ detail }) => replaceStateWithQuery('pl', detail)}
        />
      {:else}
        <span>no data</span>
      {/if}
    {:catch err}
      <pre>{err}</pre>
    {/await}
  {:else}
    {#await repoPromise}
      <Loader />
    {:then repo}
      <General {repo} />
    {:catch err}
      <pre>{err}</pre>
    {/await}
  {/if}
</div>
