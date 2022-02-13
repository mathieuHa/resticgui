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
  let repoSelected = false;
  let newRepo = false;
  let values = {
    name: '',
    entities: [{ path: '' }],
  };
  let errors = {};
  $:values = repoSelected || {
    name: '',
    entities: [{ path: '' }],
  } || newRepo;
  $:errors = (repoSelected && {
    name: '',
    entities: [{ path: '' }],
  }) || newRepo;

  const schema = yup.object?.().shape({
    name: yup.string().min(4, 'name too short').required('required'),
    entities: yup.array(yup.object({
      path: yup.string().required('required'),
    })).required('required'),
  });

  const onRefresh = async () => {
    try {
      repos = await agent.get(`/repositories`);
    } catch (err) {
      // Pass
    }
  }
  const onSubmit = async () => {
    try {
      await schema.validate(values, { abortEarly: false });
    } catch (err) {
      errors = extractErrors(err);
      console.log(errors)
    }
    console.log(values);
    try {
      await agent[newRepo ? 'post' : 'patch'](`/repositories/${newRepo ? '' : repoSelected.ID}`, values);
      repoSelected = false;
      newRepo = false;
      await onRefresh();
    } catch (err) {
      console.log(err);
    }
  }
  const onRemove = async (user) => {
    console.log(user);
    try {
      await agent.delete(`/repositories/${user.ID}`);
      await onRefresh()
    } catch (err) {
      console.log(err)
    }
  }

  onMount(onRefresh);
  console.log(values)
</script>

{#if repoSelected || newRepo}
  <Modal
    on:close={() => {
      repoSelected = false;
      newRepo = false;
    }}
  >
    <h1 class="mb-2">Update user</h1>
    <form on:submit|preventDefault|stopPropagation={onSubmit} class="flex flex-col">
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
      <div class="w-full flex justify-between">
        <label for="path">Path</label>
        <span class="text-red-500">{errors.path || ''}</span>
      </div>
      {#each values.entities || [] as entity, i}
        <input
          class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
          id="path"
          name="path"
          type="text"
          bind:value={values.entities[i].path}
        />
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          on:click|preventDefault|stopPropagation={() => values.entities.splice(1)}
        >
          -
        </button>
      {/each}
      <button
        class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
        on:click|preventDefault|stopPropagation={() => values.entities.push({ path: '' })}
      >
        +
      </button>
      <button
        class="mt-4 py-1 rounded transition bg-blue-500 hover:bg-blue-600"
        type="submit">Submit</button>
    </form>
  </Modal>
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each repos as repo, i}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p>{repo.name}</p>
      <p>added {dayjs(repo.created_at).format('D/MM/YY')}</p>

      <div class="flex mt-auto pt-2 h-8 mb-1">
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          on:click|preventDefault|stopPropagation={() => repoSelected = Object.assign({}, repo)}
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
    on:click={() => newRepo = true}
    class="text-3xl p-6 max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center flex justify-center cursor-pointer"
  >
    +
  </div>
</div>
