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
{#if values.running}
    <span class="bg-green-600 rounded-full w-30 min-w-min text-center">running status</span>
{:else}
    <div class="bg-red-600 rounded-full w-30 min-w-min text-center">running status</div>
{/if}
