<script>
  import { createEventDispatcher } from 'svelte';
  import Input from '../Input.svelte';

  export let initialPruneSchedule = {};
  let values = initialPruneSchedule;

  const dispatch = createEventDispatcher();

  $:dispatch('update', values);
</script>

<Input
  id="cronSyntax"
  type="text"
  label="Cron syntax"
  placeholder="*/30 * * * *"
  bind:value={values.cronSyntax}
  on:change={() => { values = { choice: 'cron', cronSyntax: values.cronSyntax, interval: '' }; }}
/>
<Input
  id="interval"
  type="text"
  label="Interval"
  placeholder="1h30"
  bind:value={values.interval}
  on:change={() => { values = { choice: 'interval', interval: values.interval, cronSyntax: '' }; }}
/>
