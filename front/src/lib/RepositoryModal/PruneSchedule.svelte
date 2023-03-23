<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { PruneSchedule } from '$utils/types';

  export let initialValues: PruneSchedule;
  let values = initialValues;

  const dispatch = createEventDispatcher();
  $:dispatch('update', values);
</script>

<div class="form-control">
  <label for="cronSyntax" class="label">
    <span class="label-text">Cron syntax</span> 
  </label>
  <input
    id="cronSyntax"
    type="text"
    class="input input-bordered w-full"
    placeholder="*/30 * * * *"
    bind:value={values.cronSyntax}
    on:change={() => { values = { choice: 'cron', cronSyntax: values.cronSyntax, interval: '' }; }}
  />
</div>
<div class="form-control">
  <label for="interval" class="label">
    <span class="label-text">Interval</span> 
  </label>
  <input
    id="interval"
    type="text"
    class="input input-bordered w-full"
    placeholder="1h30"
    bind:value={values.interval}
    on:change={() => { values = { choice: 'interval', interval: values.interval, cronSyntax: '' }; }}
  />
</div>