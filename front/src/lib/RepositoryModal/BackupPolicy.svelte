<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Input from '../Input.svelte';
  import type { BackupPolicy } from '$utils/types';

  export let initialValues: BackupPolicy;
  let values = initialValues;

  const dispatch = createEventDispatcher();

  $:dispatch('update', {
    ...values,
    keepAll: values.nbLastBackup === 0,
    nbLastBackup: values.nbLastBackup && values.nbLastBackup > 0 ? values.nbLastBackup : null,
  });
</script>

<Input
  id="nbLastBackup"
  name="nbLastBackup"
  type="number"
  label="Number last backup (if 0, keep all)"
  min="0"
  bind:value={values.nbLastBackup}
/>
<Input
  id="oneByHour"
  name="oneByHour"
  type="number"
  min="0"
  label="Keep One by hour for"
  bind:value={values.oneByHour}
/>
<Input
  id="oneByDay"
  name="oneByDay"
  type="number"
  min="0"
  label="Keep One by day for"
  bind:value={values.oneByDay}
/>
<Input
  id="oneByMonth"
  name="oneByMonth"
  type="number"
  min="0"
  label="Keep One by month for"
  bind:value={values.oneByMonth}
/>
