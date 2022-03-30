<script>
  import { createEventDispatcher, onMount } from 'svelte';

  import agent from '../utils/api';

  import Modal from './Modal.svelte';

  export let repositoryId = null;

  let step = 0;
  let values = null;
  let link = null;
  let logs = [];

  const dispatch = createEventDispatcher();
  
  onMount(async () => {
    const data = await agent.get(`/repository/${repositoryId}/logs`);
    logs = data.map((log) => {
      const date = new Date(log.startTime).toISOString().split('T');
      return {
        ...log,
        __date: date[0],
        __time: date[1],
      };
    });
  })
</script>

<Modal on:close={() => dispatch('close')} title="Logs for repository {repositoryId}">
  <div class="p-3">
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
  </div>
</Modal>
  