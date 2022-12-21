<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let value: string | null | undefined = '';
  export let checked: boolean | null | undefined = false;
  export let id = '';
  export let label = '';
  export let type = 'text';
  export let options = [];
  export let selected: string | null = '';
  export let secondary = false;

  let idCheckbox = id;
  let idNumber = id;
  let idPassword = id;
  let val = type === 'text' ? value : null;
  let select = type === 'select' ? selected : null;
  let check = type === 'checkbox' ? checked : null;

  const dispatch = createEventDispatcher();

  $:val = value;
</script>

{#if type === 'text'}
  <div class="flex flex-col mb-2">
    <label for={id}>{label}</label>
    <input
      class="px-2 py-1 focus:shadow-sm shadow-gray-300 rounded
        bg-gray-{secondary ? 300 : 200} dark:bg-gray-{secondary ? 800 : 700}
        disabled:bg-gray-700/[.5] dark:disabled:bg-gray-200/[.5] transition-all"
      id={id}
      type="text"
      bind:value={val}
      on:change={() => { value = val; dispatch('change'); }}
      {...$$props}
    />
  </div>
{:else if type === 'password'}
  <div class="flex flex-col mb-2">
    <label for={idPassword}>{label}</label>
    <input
      class="px-2 py-1 focus:shadow-sm shadow-gray-300 rounded
        bg-gray-{secondary ? 300 : 200} dark:bg-gray-{secondary ? 800 : 700}
        disabled:bg-gray-700/[.5] dark:disabled:bg-gray-200/[.5] transition-all"
      id={idPassword}
      type="password"
      bind:value={val}
      on:change={() => { value = val; dispatch('change'); }}
      {...$$props}
    />
  </div>
{:else if type === 'select'}
  <div class="flex flex-col mb-2">
    <label for="select">{label}</label>
    <select
      class="px-2 py-1 focus:shadow-sm shadow-gray-300 rounded
        bg-gray-{secondary ? 300 : 200} dark:bg-gray-{secondary ? 800 : 700}
        disabled:bg-gray-700/[.5] dark:disabled:bg-gray-200/[.5] transition-all"
      id="select" 
      bind:value={select}
      on:change={() => { selected = select; dispatch('change'); }}
    >
      <option value="" default>{$$props.placeholder || 'No selected'}</option>
      {#each options as option}
        <option value={option}>{option}</option>
      {/each}
    </select>
  </div>
{:else if type === 'number'}
  <div class="flex flex-col mb-2">
    <label for={idNumber}>{label}</label>
    <input
      class="px-2 py-1 focus:shadow-sm shadow-gray-300 rounded
        bg-gray-{secondary ? 300 : 200} dark:bg-gray-{secondary ? 800 : 700}
        disabled:bg-gray-700/[.5] dark:disabled:bg-gray-200/[.5] transition-all"
      id={idNumber}
      type="number"
      min="0"
      bind:value={val}
      on:change={() => { value = val; dispatch('change'); }}
      {...$$props}
    />
  </div>
{:else if type === 'checkbox'}
  <label class="inline-flex mr-auto items-center cursor-pointer" for={id}>
    <span class="relative block w-[30px] h-[15px] mt-1">
      <input
        class="hidden opacity-0"
        id={idCheckbox}
        name={idCheckbox}
        type="checkbox"
        bind:checked={check}
        on:change={() => { checked = check; dispatch('change'); }}
      />
      <span class="slider absolute top-0 left-0 right-0 bottom-0 bg-gray-300 rounded-lg transition-all"></span>
    </span>
    <span class="ml-2 leading-8">{label}</span>
  </label>
{/if}

<style>
  .slider:before {
    position: absolute;
    content: "";
    height: calc(100% - 4px);
    width: calc(50% - 4px);
    left: 2px;
    bottom: 2px;
    background-color: white;
    border-radius: 50%;
    transition: left 250ms;
  }
  input:checked + .slider {
    background-color: #2196f3;
  }
  input:checked + .slider:before {
    left: calc(50% + 2px);
  }
</style>
