<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { Repository } from '$utils/types';

  export let initialValues: Repository;

  let values = {
    name: initialValues.name,
    password: initialValues.password,
    active: initialValues.active,
    pruneActive: initialValues.pruneActive,
  };

  const dispatch = createEventDispatcher();

  $:dispatch('update', values);
  $:console.log(values)
</script>

<div class="form-control">
  <label for="name" class="label">
    <span class="label-text">Name</span> 
  </label>
  <input id="name" type="text" bind:value={values.name} class="input input-bordered w-full " />
</div>
<div class="form-control">
  <label for="password" class="label">
    <span class="label-text">Password</span> 
  </label>
  <input
    id="password"
    type="password"
    disabled={!!initialValues.id}
    bind:value={values.password}
    class="input input-bordered w-full " />
</div>
<div class="form-control">
  <label class="label cursor-pointer">
    <span class="label-text">Enable backup</span>
    <input type="checkbox" bind:checked={values.active} class="checkbox" />
  </label>
</div>
<div class="form-control">
  <label class="label cursor-pointer">
    <span class="label-text">Enable prune</span>
    <input type="checkbox" bind:checked={values.pruneActive} class="checkbox"  />
  </label>
</div>
