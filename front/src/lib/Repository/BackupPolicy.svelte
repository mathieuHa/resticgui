<script>
  import { createEventDispatcher } from 'svelte';
  import Input from '../Input.svelte';

  export let initialBackupPolicy = {};
  let values = initialBackupPolicy;

  const dispatch = createEventDispatcher();

  $:dispatch('update', {
    ...values,
    keepAll: values.nbLastBackup === 0,
    nbLastBackup: values.nbLastBackup > 0 ? values.nbLastBackup : null,
  });
  $:console.log(values)
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
  type="checkbox"
  label="One by hour"
  bind:checked={values.oneByHour}
/>
<Input
  id="oneByDay"
  name="oneByDay"
  type="checkbox"
  label="One by day"
  bind:checked={values.oneByDay}
/>
<Input
  id="oneByMonth"
  name="oneByMonth"
  type="checkbox"
  label="One by month"
  bind:checked={values.oneByMonth}
/>
