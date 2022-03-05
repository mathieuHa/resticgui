<script>
  import { createEventDispatcher } from 'svelte';
  import Input from '../Input.svelte';

  export let initialValues = {};

  let values = initialValues;

  const dispatch = createEventDispatcher();

  $:dispatch('update', values);
</script>

<Input
  id="name"
  label="Name"
  bind:value={values.name}
/>
<Input
  id="active"
  label="Enable backup"
  type="checkbox"
  bind:checked={values.active}
/>
