<script>
  import { page } from '$app/stores';

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
      logs = logs.map((log) => {
        const date = new Date(log.startTime).toISOString().split('T');
        return {
          ...log,
          __date: date[0],
          __time: date[1],
        };
      });
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

  $: console.log(tab, repo, stats);
  $: onRefresh($page.params.id);
</script>

<div class="w-full max-w-full">
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
      {#each snapshots as snapshot}
        <pre class="text-sm">{JSON.stringify(snapshot, null, 1)}</pre>
      {/each}
    {:else if tab === 'logs'}
      {#each logs as log, i}
        {#if i === 0 || log.__date !== logs[i - 1].__date}
          <p class="text-lg font-bold">{log.__date}</p>
        {/if}
        <div class="pl-5 pt-2">
          <p><span class="underline">{log.__time.slice(0, 8)}</span> - {log.operation}</p>
          <div class="pt-1 pl-5 flex flex-col text-sm">
            {#if log.error}
              <pre><span class="text-red-500">err:</span> {log.error}</pre>
            {/if}
            {#if log.stderr}
              <pre><span class="text-red-500">stderr:</span> {log.stderr}</pre>
            {/if}
            {#if log.stdout}
              <pre><span class="text-blue-600 dark:text-blue-300">stdout:</span> {log.stdout}</pre>
            {/if}
          </div>
        </div>
      {/each}
    {:else if tab === 'stats'}
      <pre>{stats}</pre>
    {:else}
      <pre class="text-sm">{JSON.stringify(repo, null, 4)}</pre>
    {/if}
  {/if}
</div>