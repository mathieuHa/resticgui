<script>
  import { createEventDispatcher } from 'svelte';

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

<label for="name">File name</label>
<input
  class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
  id="name"
  name="name"
  type="text"
  bind:value={values.name}
/>
<label for="nbLastBackup">Number last backup <span class="text-sm">(if 0, keep all)</span></label>
<input
  class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
  id="nbLastBackup"
  name="nbLastBackup"
  type="number"
  min="0"
  bind:value={values.nbLastBackup}
/>
<div>
  <input
    class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
    id="oneByHour"
    name="oneByHour"
    type="checkbox"
    bind:checked={values.oneByHour}
  />
  <label for="oneByHour">One by hour</label>
</div>
<div>
  <input
    class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
    id="oneByDay"
    name="oneByDay"
    type="checkbox"
    bind:checked={values.oneByDay}
  />
  <label for="oneByDay">One by day</label>
</div>
<div>
  <input
    class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
    id="oneByMonth"
    name="oneByMonth"
    type="checkbox"
    bind:checked={values.oneByMonth}
  />
  <label for="oneByMonth">One by month</label>
</div>
