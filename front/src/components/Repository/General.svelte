<script>
  import { createEventDispatcher } from 'svelte';

  export let initialValues = {};

  let values = initialValues;

  const dispatch = createEventDispatcher();

  $:dispatch('update', values);
</script>

<label for="name">Name</label>
<input
  class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
  id="name"
  name="name"
  type="text"
  bind:value={values.name}
/>
<div>
  <input
    class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
    id="active"
    name="active"
    type="checkbox"
    bind:checked={values.active}
  />
  <label for="active">Enable backup</label>
</div>
