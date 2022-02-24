<script>
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';
  import * as yup from 'yup';

  import agent from '../utils/api';
  import extractErrors from '../utils/extractErrors';
  import Pen from '../components/Icon/Pen.svelte';
  import Trash from '../components/Icon/Trash.svelte';
  import Modal from '../components/Modal.svelte';

  let users = [];
  let errors = {};
  let values = null;

  const schema = yup.object?.().shape({
    firstname: yup.string().required('required'),
    lastname: yup.string().required('required'),
    email: yup.string().required('required').email('not an email'),
  });

  const onRefresh = async () => {
    try {
      users = await agent.get(`/users`);
    } catch (err) {
      // Pass
    }
  }
  const onSubmit = async () => {
    try {
      await schema.validate(values, { abortEarly: false });
    } catch (err) {
      errors = extractErrors(err);
    }
    console.log(values);
    try {
      await agent[values.ID ? 'patch' : 'post'](`/users/${values.ID || ''}`, values);
      await onRefresh();
      values = null;
    } catch (err) {
      // Pass
    }
  }
  const onRemove = async (user) => {
    console.log(user);
    try {
      await agent.delete(`/users/${user.ID}`);
      await onRefresh()
    } catch (err) {
      console.log(err)
    }
  }

  onMount(onRefresh);
</script>

{#if values}
  <Modal on:close={() => { values = null; }}>
    <h1 class="mb-2">Update user</h1>
    <form on:submit|preventDefault|stopPropagation={onSubmit} class="flex flex-col">
      <div class="w-full flex justify-between">
        <label for="firstname">Firstname</label>
        <span class="text-red-500">{errors.firstname || ''}</span>
      </div>
      <input
        class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
        id="firstname"
        name="firstname"
        type="text"
        bind:value={values.firstname}
      />
      <div class="w-full flex justify-between">
        <label for="lastname">Lastname</label>
        <span class="text-red-500">{errors.lastname || ''}</span>
      </div>
      <input
        class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
        id="lastname"
        name="lastname"
        type="text"
        bind:value={values.lastname}
      />
      <div class="w-full flex justify-between">
        <label for="email">Email</label>
        <span class="text-red-500">{errors.email || ''}</span>
      </div>
      <input
        class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
        id="email"
        name="email"
        type="text"
        bind:value={values.email}
      />
      <button
        class="mt-4 py-1 rounded transition bg-blue-500 hover:bg-blue-600"
        type="submit"
      >
        Submit
      </button>
    </form>
  </Modal>
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each users as user, i}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p class="truncate text-lg font-bold mb-2">{user.firstname} {user.lastname}</p>
      <p class="truncate" title={user.email}>{user.email}</p>
      <p class="truncate text-sm my-2 font-light">{dayjs(user.createdAt).format('D/MM/YY')}</p>

      <div class="flex mt-auto pt-2 h-8 mb-1">
        <button
          class="rounded h-6 w-6text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          on:click|preventDefault|stopPropagation={() => { values = Object.assign({}, user); }}
        >
          <Pen />
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 ml-2 transition bg-orange-500 hover:bg-orange-600"
          on:click|preventDefault|stopPropagation={() => onRemove(user)}
        >
          <Trash />
        </button>
      </div>
    </div>
  {/each}
  <div
    on:click={() => { values = { firstname: '', lastname: '', email: '' }; }}
    class="text-3xl p-6 max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center flex justify-center cursor-pointer"
  >
    +
  </div>
</div>
