<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { BackupPolicy } from '$utils/types';

  export let initialValues: BackupPolicy;
  let values = initialValues;

  const dispatch = createEventDispatcher();

  $:dispatch('update', {
    ...values,
    //keepAll: values.nbLastBackup === 0,
    //nbLastBackup: values.nbLastBackup && values.nbLastBackup > 0 ? values.nbLastBackup : null,
  });
  console.log(values);
</script>

<div class="form-control">
  <label class="label cursor-pointer">
    <span class="label-text">Keep All backup</span>
    <input type="checkbox" bind:checked={values.keepAll} class="checkbox"  />
  </label>
</div>
<div class="form-control">
  <label for="nbKeepLast" class="label">
    <span class="label-text">Number last backup</span> 
  </label>
  <input
    id="nbKeepLast"
    name="nbKeepLast"
    type="number"
    min="0"
    disabled={values.keepAll}
    bind:value={values.nbKeepLast}
    class="input input-bordered w-full"
  />
</div>
<div class="form-control">
  <label for="nbKeepHours" class="label">
    <span class="label-text">Keep One by hour for</span> 
  </label>
  <input
    id="nbKeepHours"
    type="number"
    min="0"
    disabled={values.keepAll}
    bind:value={values.nbKeepHours}
    class="input input-bordered w-full"
  />
</div>
<div class="form-control">
  <label for="nbKeepDays" class="label">
    <span class="label-text">Keep One by day for</span> 
  </label>
  <input
    id="nbKeepDays"
    type="number"
    min="0"
    disabled={values.keepAll}
    bind:value={values.nbKeepDays}
    class="input input-bordered w-full"
  />
</div>
<div class="form-control">
  <label for="nbKeepWeeks" class="label">
    <span class="label-text">Keep One by week for</span> 
  </label>
  <input
    id="nbKeepWeeks"
    type="number"
    min="0"
    disabled={values.keepAll}
    bind:value={values.nbKeepWeeks}
    class="input input-bordered w-full"
  />
</div>
<div class="form-control">
  <label for="nbKeepMonths" class="label">
    <span class="label-text">Keep One by month for</span> 
  </label>
  <input
    id="nbKeepMonths"
    type="number"
    disabled={values.keepAll}
    bind:value={values.nbKeepMonths}
    class="input input-bordered w-full"
    min="0"
  />
</div>
