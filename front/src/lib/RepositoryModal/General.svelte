<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Input from '$lib/Input.svelte';
  import type { Repository } from '$utils/types';

  export let initialValues: Repository;
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
  id="password"
  label="Password (cannot be changed later)"
  type="password"
  bind:value={values.password}
  disabled={!!initialValues.id}
/>
<Input
  id="active"
  label="Enable backup"
  type="checkbox"
  bind:checked={values.active}
/>
<Input
  id="pruneActive"
  label="Enable prune"
  type="checkbox"
  bind:checked={values.pruneActive}
/>
