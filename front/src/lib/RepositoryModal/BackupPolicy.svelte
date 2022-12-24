<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Input from '../Input.svelte';
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
<Input
  id="keepAll"
  label="Keep All backup"
  type="checkbox"
  bind:checked={values.keepAll}
/>
<Input
  id="nbKeepLast"
  name="nbKeepLast"
  type="number"
  label="Number last backup"
  min="0"
  disabled={values.keepAll}
  bind:value={values.nbKeepLast}
/>
<Input
  id="nbKeepHours"
  name="nbKeepHours"
  type="number"
  min="0"
  label="Keep One by hour for"
  disabled={values.keepAll}
  bind:value={values.nbKeepHours}
/>
<Input
  id="nbKeepDays"
  name="nbKeepDays"
  type="number"
  min="0"
  label="Keep One by day for"
  disabled={values.keepAll}
  bind:value={values.nbKeepDays}
/>
<Input
  id="nbKeepWeeks"
  name="nbKeepWeeks"
  type="number"
  min="0"
  label="Keep One by week for"
  disabled={values.keepAll}
  bind:value={values.nbKeepWeeks}
/>
<Input
  id="nbKeepMonths"
  name="nbKeepMonths"
  type="number"
  min="0"
  label="Keep One by month for"
  disabled={values.keepAll}
  bind:value={values.nbKeepMonths}
/>
