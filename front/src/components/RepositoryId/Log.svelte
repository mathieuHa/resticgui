<script>
  import Input from '../Input.svelte';

  export let logs;
  
  let displayedLogs = [];
  let filters = {};
  let filter = '';
  
  $: displayedLogs = logs.map((log) => {
    filters[log.operation] = true;
    const date = new Date(log.startTime).toISOString().split('T');
    return {
      ...log,
      __date: date[0],
      __time: date[1],
    };
  });
</script>

<Input
  type="select"
  label="Filter logs"
  placeholder="All"
  options={Object.keys(filters)}
  bind:selected={filter}
/>
{#each displayedLogs as log, i}
  {#if !filter || log.operation === filter}
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
  {/if}
{/each}