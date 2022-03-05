<script>
  import { createEventDispatcher } from 'svelte';

  export let value = '';
  export let checked = false;
  export let id = '';
  export let label = '';
  export let type = 'text';
  export let placeholder = '';

  let idCheckbox = id;
  let idNumber = id;
  let val = type === 'text' ? value : null;
  let check = type === 'checkbox' ? checked : null;

  const dispatch = createEventDispatcher();

  $:val = value;
</script>

{#if type === 'text'}
  <div class="flex flex-col mb-2">
    <label for={id}>{label}</label>
    <input
      class="px-2 py-1 focus:shadow-sm shadow-gray-300 dark:shadow-gray-700 rounded bg-gray-200 dark:bg-gray-700"
      id={id}
      type="text"
      placeholder={placeholder}
      bind:value={val}
      on:change={() => { value = val; dispatch('change'); }}
    />
  </div>
{:else if type === 'number'}
  <div class="flex flex-col mb-2">
    <label for={idNumber}>{label}</label>
    <input
      class="px-2 py-1 focus:shadow-sm shadow-gray-300 dark:shadow-gray-700 rounded bg-gray-200 dark:bg-gray-700"
      id={idNumber}
      type="number"
      min="0"
      bind:value={val}
      on:change={() => { value = val; dispatch('change'); }}
    />
  </div>
{:else if type === 'checkbox'}
  <div>
    <input
      class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700 cursor-pointer"
      id={idCheckbox}
      name={idCheckbox}
      type="checkbox"
      bind:checked={check}
      on:change={() => { checked = check; dispatch('change'); }}
    />
    <label class="cursor-pointer" for={id}>{label}</label>
  </div>
{/if}
