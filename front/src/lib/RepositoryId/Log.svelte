<script lang="ts">
  interface Log {
    __date?: string
    __time?: string
    startTime: string
    endTime: string
    id: number
    operation: string
    stdout: string
    stderr: string
    error: string
  }

  export let logs: Log[];
  
  let displayedLogs: Log[] = [];
  let filters: Record<string, boolean> = {};
  let filter = '';
  let seeMore: Record<number, boolean> = {}
  
  $: displayedLogs = logs.map((log) => {
    filters[log.operation] = true;
    const date = log.startTime.split('T');
    return {
      ...log,
      __date: date[0],
      __time: date[1],
    };
  });
</script>

{#if displayedLogs.length > 0}
  <select class="select w-full mb-5" bind:value={filter}>
    <option value="" selected>All</option>
    {#each Object.keys(filters) as option}
      <option>{option}</option>
    {/each}
  </select>
  {#each displayedLogs as log, i}
    {#if !filter || log.operation === filter}
      {#if i === 0 || log.__date === logs[i - 1]?.__date}
        <p class="text-lg font-bold">{log.__date}</p>
      {/if}
      <div class="pl-5 pt-2">
        <p><span class="underline">{log.__time?.slice(0, 8)}</span> - {log.operation}</p>
        <div class="pt-1 pl-5 flex flex-col text-sm whitespace-normal">
          {#if log.error}
            <pre class="inline break-words whitespace-normal"><span class="text-error">err:</span> {log.error}</pre>
          {/if}
          {#if log.stderr}
            <pre class="inline break-words vwhitespace-normal"><span class="text-error">stderr:</span> {log.stderr}</pre>
          {/if}
          {#if log.stdout}
            <pre class="inline break-words whitespace-normal"><span class="text-info">stdout:</span> {seeMore[log.id] ? log.stdout : `${log.stdout.substring(0, 100)}.....`}</pre>
            {#if log.stdout.length > 100 && !seeMore[log.id]}
              <span>
                <button class="btn btn-xs mt-1" on:click={() => (seeMore[log.id] = true)}>See more</button>
              </span>
            {/if}
          {/if}
        </div>
      </div>
    {/if}
  {/each}
{/if}