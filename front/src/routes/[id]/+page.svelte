<script lang="ts">
  import { onDestroy } from 'svelte';

  import { page } from '$app/stores';

  import Log from '$lib/RepositoryId/Log.svelte';
  import Snapshot from '$lib/RepositoryId/Snapshots/Snapshot.svelte';
  import General from '$lib/RepositoryId/General.svelte';
  import Loader from '$lib/Loader.svelte';
  import agent from '$utils/api';
  import { refresh } from '$utils/stores';

  let tab = new URLSearchParams($page.url.search).get('tab') || 'general';

  const fetch = async (path: string) => {
    return await agent.get(path);
  };

  const replaceStateWithQuery = (value: string) => {
    const url = new URL(window.location.toString());
    url.searchParams.set('tab', value);
    history.replaceState({}, '', url);
    tab = value;
  };

  let repoPromise = fetch(`/repositories/${$page.params.id}`);
  let logsPromise = fetch(`/repositories/${$page.params.id}/logs`);
  let snapshotsPromise = fetch(`/repositories/${$page.params.id}/snapshots`);
  const unsuscribe = refresh.subscribe(() => {
    repoPromise = fetch(`/repositories/${$page.params.id}`);
    logsPromise = fetch(`/repositories/${$page.params.id}/logs`);
    snapshotsPromise = fetch(`/repositories/${$page.params.id}/snapshots`);
  });
  onDestroy(unsuscribe);
</script>

<div class="mt-2 mb-10 w-full max-w-full">
  <div class="flex justify-between flex-wrap mb-5">
    {#each ["General", "Snapshots", "Logs"] as stepped, i}
      <button
        on:click={() => replaceStateWithQuery(stepped.toLowerCase())}
        class="flex-1 pb-2 px-4 {tab === stepped.toLowerCase()
          ? 'border-b-2 border-gray-500 order-4 md:order-unset'
          : 'cursor-pointer hover:border-b-2 hover:border-b-blue-500/[.5]'}"
      >
        <span class="rounded-xl">{stepped}</span>
      </button>
    {/each}
  </div>

  {#if tab === 'snapshots'}
    {#await snapshotsPromise}
      <Loader />
    {:then snapshots}
      {#if snapshots && snapshots.length > 0}
        <Snapshot 
          on:delete={() => (snapshotsPromise = fetch(`/repositories/${$page.params.id}/snapshots`))}
          snapshots={snapshots.sort((a, b) => b.time.localeCompare(a.time))}
        />
      {:else}
        <span>no data</span>
      {/if}
    {:catch err}
      <pre>{err}</pre>
    {/await}
  {:else if tab === 'logs'}
    {#await logsPromise}
      <Loader />
    {:then logs}
      {#if logs && logs.length > 0}
        <Log {logs} />
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
