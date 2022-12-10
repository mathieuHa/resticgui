<script>
  import { page } from '$app/stores';

  import Log from '../../components/RepositoryId/Log.svelte';
  import Snapshot from '../../components/RepositoryId/Snapshot.svelte';
  import General from '../../components/RepositoryId/General.svelte';
  import agent from '../../utils/api';
  import { toasts } from '../../utils/stores';

  let repo = null;
  let stats = null;
  let snapshots = [];
  let logs = [];
  let tab = new URLSearchParams($page.url.search).get('tab') || null;

  const fetch = async (path) => {
    try {
      return await agent.get(path);
    } catch (err) {
      return;
    }
  }
  const onRefresh = async (id) => {
    try {
      [repo, stats, logs, snapshots] = await Promise.all([
        fetch(`/repositories/${id}`),
        fetch(`/repositories/${id}/stats`),
        fetch(`/repositories/${id}/logs`),
        fetch(`/repositories/${id}/snapshots`),
      ]);
    } catch (err) {
      toasts.addToast(err?.message, 'failure');
    }
  }

  const replaceStateWithQuery = (value) => {
    const url = new URL(window.location.toString());
    url.searchParams.set('tab', value);
    history.replaceState({}, '', url);
    tab = value;
  };

  $: onRefresh($page.params.id);
</script>

<div class="mt-2 mb-10 w-full max-w-full">
  <div class="flex justify-between flex-wrap mb-3">
    {#each ['General', 'Snapshots', 'Logs', 'Stats'] as stepped, i}
      <div
        on:click={() => replaceStateWithQuery(stepped.toLowerCase())}
        class="flex-1 pb-2 px-4 {tab === stepped.toLowerCase() ? 'border-b-2 border-gray-500 order-4 md:order-unset' : 'cursor-pointer hover:border-b-2 hover:border-b-blue-500/[.5]'}"
      >
        <span class="rounded-xl">{stepped}</span>
      </div>
    {/each}
  </div>
  {#if repo}
    {#if tab === 'snapshots'}
      {#if logs && logs.length > 0}
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
    {:else if tab === 'stats'}
      <pre>{stats || 'no data'}</pre>
    {:else}
      <General {repo}/>
      <!-- <pre class="text-sm">{JSON.stringify(repo, null, 4)}</pre> -->
    {/if}
  {/if}
</div>