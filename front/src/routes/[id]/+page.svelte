<script lang="ts">
  import { page } from '$app/stores';

  import Log from '$lib/RepositoryId/Log.svelte';
  import Snapshot from '$lib/RepositoryId/Snapshot.svelte';
  import General from '$lib/RepositoryId/General.svelte';
  import Loader from '$lib/Loader.svelte';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';

  let repo = null;
  let snapshots = null;
  let logs = null;
  let tab = new URLSearchParams($page.url.search).get('tab') || null;
  let loading = true;

  const fetch = async (path: string) => {
    try {
      return await agent.get(path);
    } catch (err) {
      return null;
    }
  };

  const onRefresh = async (id: string) => {
    try {
      [repo, logs, snapshots] = await Promise.all([
        fetch(`/repositories/${id}`),
        fetch(`/repositories/${id}/logs`),
        fetch(`/repositories/${id}/snapshots`),
      ]);
    } catch (err: unknown) {
      toasts.addToast(err?.message, 'failure');
    }
    loading = false;
  };

  const replaceStateWithQuery = (value: string) => {
    const url = new URL(window.location.toString());
    url.searchParams.set('tab', value);
    history.replaceState({}, '', url);
    tab = value;
  };

  $: onRefresh($page.params.id as string);
</script>

<div class="mt-2 mb-10 w-full max-w-full">
  <div class="flex justify-between flex-wrap mb-3">
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

  {#if loading}
    <Loader />
  {:else}
    {#if repo}
      {#if tab === 'snapshots'}
        {#if snapshots && snapshots.length > 0}
          <Snapshot snapshots={snapshots.sort((a, b) => b.time.localeCompare(a.time))} />
        {:else}
          <span>no data</span>
        {/if}
      {:else if tab === 'logs'}
        {#if logs && logs.length > 0}
          <Log {logs} />
        {:else}
          <span>no data</span>
        {/if}
      {:else}
        <General {repo} />
      {/if}
    {/if}
  {/if}
</div>
