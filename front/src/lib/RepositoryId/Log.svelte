<script lang="ts">
  import Input from '../Input.svelte';

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
  <Input
    type="select"
    label="Filter logs"
    placeholder="All"
    options={Object.keys(filters)}
    bind:selected={filter}
  />
  {#each displayedLogs as log, i}
    {#if !filter || log.operation === filter}
      {#if i === 0 || log.__date === logs[i - 1]?.__date}
        <p class="text-lg font-bold">{log.__date}</p>
      {/if}
      <div class="pl-5 pt-2">
        <p><span class="underline">{log.__time?.slice(0, 8)}</span> - {log.operation}</p>
        <div class="pt-1 pl-5 flex flex-col text-sm">
          {#if log.error}
            <pre><span class="text-red-500">err:</span> {log.error}</pre>
          {/if}
          {#if log.stderr}
            <pre><span class="text-red-500">stderr:</span> {log.stderr}</pre>
          {/if}
          {#if log.stdout}
            <pre><span class="text-blue-600 dark:text-blue-300">stdout:</span> {seeMore[log.id] ? log.stdout : `${log.stdout.substring(0, 100)}.....`}</pre>
            {#if log.stdout.length > 100}
              <span>
                <button
                  class="bg-blue-500 hover:bg-blue-600 px-1 inline rounded mt-1"
                  on:click={() => (seeMore[log.id] = true)}
                >
                  See more
                </button>
              </span>
            {/if}
          {/if}
        </div>
      </div>
    {/if}
  {/each}
{/if}