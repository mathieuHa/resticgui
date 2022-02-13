<script>
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';
  import * as yup from 'yup';

  import extractErrors from '../utils/extractErrors';
  import Pen from '../components/icon/Pen.svelte';
  import Trash from '../components/icon/Trash.svelte';
  import Modal from '../components/Modal.svelte';

  let users = [];
  let userSelected = false;
  let newUser = false;
  let values = {};
  let errors = {};
  $:values = userSelected || {} || newUser;
  $:errors = (userSelected && {}) || newUser;

  const schema = yup.object?.().shape({
    firstname: yup.string().required('required'),
    lastname: yup.string().required('required'),
    email: yup.string().required('required').email('not an email'),
  });

  const onRefresh = async () => {
    try {
      console.log('wesh')
      users = await (await fetch(`/api/v1/users`)).json();
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
      await fetch(`/api/v1/users/${newUser ? '' : userSelected.ID}`, {
        method: newUser ? 'POST' : 'PATCH',
        body: JSON.stringify(values),
      });
      await onRefresh();
      userSelected = false;
      newUser = false;
    } catch (err) {
      // Pass
    }
  }
  const onRemove = async (user) => {
    console.log(user);
    try {
      await fetch(`/api/v1/users/${user.ID}`, {
        method: 'DELETE',
      });
      await onRefresh()
    } catch (err) {
      // Pass
    }
  }

  onMount(onRefresh);

</script>

{#if userSelected || newUser}
  <Modal
    on:close={() => {
      userSelected = false;
      newUser = false;
    }}
  >
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
        type="submit">Submit</button>
    </form>
  </Modal>
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each users as user, i}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p>{user.firstname} {user.lastname}</p>
      <p>{user.email}</p>
      <p>added {dayjs(user.CreatedAt).format('D/MM/YY')}</p>

      <div class="flex mt-auto pt-2 h-8 mb-1">
        <button
          class="rounded h-6 w-6text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          on:click|preventDefault|stopPropagation={() => userSelected = Object.assign({}, user)}
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
    on:click={() => newUser = true}
    class="text-3xl p-6 max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center flex justify-center cursor-pointer"
  >
    +
  </div>
</div>
