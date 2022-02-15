<script>
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';
  import * as yup from 'yup';

  import agent from '../utils/api';
  import extractErrors from '../utils/extractErrors';
  import Pen from '../components/icon/Pen.svelte';
  import Trash from '../components/icon/Trash.svelte';
  import Modal from '../components/Modal.svelte';

  let repos = [];
  let errors = {};
  let values = null;
  let step = 0;

  const schema = yup.object?.().shape({
    name: yup.string().min(4, 'too short').required('required'),
    includedPaths: yup.array(yup.object().shape({
      name: yup.string().min(1, 'too short').required('required'),
    })).min(1, 'too short').required('required'),
    excludedPaths: yup.array(yup.object().shape({
      name: yup.string().min(1, 'too short').required('required'),
    })).min(1, 'too short').required('required'),
  });

  const onRefresh = async () => {
    try {
      repos = await agent.get(`/repositories`);
    } catch (err) {
      // Pass
    }
  }
  const onSubmit = async () => {
    if (step !== 2) {
      step += 1;
      return;
    }
    try {
      await schema.validate(values, { abortEarly: false });
    } catch (err) {
      errors = extractErrors(err);
      console.log(errors)
    }
    console.log(values);
    try {
      await agent[values.ID ? 'patch' : 'post'](`/repositories/${values.ID || ''}`, values);
      await onRefresh();
      values = null;
    } catch (err) {
      console.log(err);
    }
  }
  const onRemove = async (repo) => {
    try {
      await agent.delete(`/repositories/${repo.ID}`);
      await onRefresh()
    } catch (err) {
      console.log(err)
    }
  }

  $:if (values === null) step = 0;

  onMount(onRefresh);
</script>

{#if values}
  <Modal on:close={() => { values = null; }}>
    <div class="flex max-w-min m-auto justify-between">
      {#each [0,1,2] as stepped}
        <div
          on:click={() => { step = stepped; }}
          class={`w-3 h-3 mx-2 rounded-xl cursor-pointer bg-${step === stepped ? 'orange' : 'blue'}-500`}
        ></div>
      {/each}
    </div>
    <h1 class="mb-2">Update repository</h1>
    <form on:submit|preventDefault|stopPropagation={onSubmit} class="flex flex-col">
      <div class={`flex flex-col ${step !== 0 && 'hidden'}`}>
        <div class="w-full flex justify-between">
          <label for="name">Name</label>
          <span class="text-red-500">{errors.name || ''}</span>
        </div>
        <input
          class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
          id="name"
          name="name"
          type="text"
          bind:value={values.name}
        />
      </div>
      <div class={`flex flex-col ${step !== 1 && 'hidden'}`}>
        <div class="w-full flex justify-between">
          <label>Paths</label>
          <span class="text-red-500">{errors.path || ''}</span>
        </div>
        {#each values.includedPaths as path, i}
          <div class="flex py-2 items-center">
            <input
              class="px-2 flex-1 py-1 mr-2 border-2 border-indigo-900 rounded dark:bg-gray-700"
              id={`path_${i}`}
              name={`path_${i}`}
              type="text"
              bind:value={values.includedPaths[i].name}
            />
            <div
              class="rounded h-full w-8 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
              on:click|self|preventDefault|stopPropagation={() => {
                console.log('caca1');
                values = Object.assign({}, { ...values, includedPaths: [
                  ...values.includedPaths.slice(0, i),
                  ...values.includedPaths.slice(i + 1),
                ] });
              }}
            >
              -
            </div>
          </div>
        {/each}
        <div
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          on:click|self|preventDefault|stopPropagation={() => {
            console.log('caca2');
            values = Object.assign({}, { ...values, includedPaths: [ ...values.includedPaths, { name: '' } ] });
          }}
        >
          +
        </div>
      </div>
      <div class={`flex flex-col ${step !== 2 && 'hidden'}`}>
        <div class="w-full flex justify-between">
          <label for="cron">Syntax Cron</label>
          <span class="text-red-500">{errors.cron || ''}</span>
        </div>
        <input
          class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
          id="cron"
          name="cron"
          type="text"
          placeholder="*/30 * * * *"
          bind:value={values.cron}
        />
        <div class="w-full flex justify-between">
          <label for="interval">Interval</label>
          <span class="text-red-500">{errors.interval || ''}</span>
        </div>
        <input
          class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
          id="interval"
          name="interval"
          type="text"
          placeholder="1h30"
          bind:value={values.interval}
        />
      </div>
      <button
        class="mt-4 py-1 rounded transition bg-blue-500 hover:bg-blue-600"
        type="submit"
      >
        {step === 2 ? 'Submit' : 'Next'}
      </button>
    </form>
  </Modal>
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each repos as repo, i}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p class="truncate text-lg font-bold">{repo.name}</p>
      <p class="truncate text-sm font-light my-2">{dayjs(repo.createdAt).format('D/MM/YY')}</p>

      <div class="flex mt-auto pt-2 h-8 mb-1">
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          on:click|preventDefault|stopPropagation={() => { values = Object.assign({}, repo); }}
        >
          <Pen />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 ml-2 transition bg-orange-500 hover:bg-orange-600"
          on:click|preventDefault|stopPropagation={() => onRemove(repo)}
        >
          <Trash />
        </button>
      </div>
    </div>
  {/each}
  <div
    on:click={() => { values = { name: '', includedPaths: [{ name: '' }] }; }}
    class="text-3xl p-6 max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center flex justify-center cursor-pointer"
  >
    +
  </div>
</div>
