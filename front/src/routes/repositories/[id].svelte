<script>
  import { page } from '$app/stores';

  import Log from '../../components/RepositoryId/Log.svelte';
  import Snapshot from '../../components/RepositoryId/Snapshot.svelte';
  import agent from '../../utils/api';

  let repo = null;
  let stats = null;
  let snapshots = [];
  let logs = [];
  let tab = new URLSearchParams($page.url.search).get('tab') || null;

  const onRefresh = async (id) => {
    try {
      [repo, stats, logs, snapshots] = await Promise.all([
        agent.get(`/repositories/${id}`),
        agent.get(`/repositories/${id}/stats`),
        agent.get(`/repositories/${id}/logs`),
        agent.get(`/repositories/${id}/snapshots`),
      ]);
    } catch (err) {
      // Pass
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

<div class="mt-2 w-full max-w-full">
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
      <Snapshot snapshots={snapshots} />
    {:else if tab === 'logs'}
      <Log logs={logs} />
    {:else if tab === 'stats'}
      <pre>{stats}</pre>
    {:else}
        <pre class="text-sm">{JSON.stringify(repo, null, 4)}</pre>
    {/if}
  {/if}
</div>