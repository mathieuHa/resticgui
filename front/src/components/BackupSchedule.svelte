<script>
  import { createEventDispatcher } from 'svelte';

  export let initialBackupSchedule = {};
  let values = initialBackupSchedule;

  const dispatch = createEventDispatcher();

  $: dispatch('update', values);
</script>

<div class="w-full flex justify-between">
  <label for="cronSyntax">Syntax Cron</label>
</div>
<input
  class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
  id="cronSyntax"
  name="cronSyntax"
  type="text"
  placeholder="*/30 * * * *"
  bind:value={values.cronSyntax}
  on:change={() => { values = { choice: 'cron', cronSyntax: values.cronSyntax } }}
/>
<div class="w-full flex justify-between">
  <label for="interval">Interval</label>
</div>
<input
  class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
  id="interval"
  name="interval"
  type="text"
  placeholder="1h30"
  bind:value={values.interval}
  on:change={() => { values = { choice: 'interval', interval: values.interval } }}
/>
