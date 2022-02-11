<script context="module">
  // export const prerender = true; // you can uncomment to prerender as an optimization
  export const hydrate = true;
  import { BACKEND_URL } from '$lib/siteConfig';
  export async function load({ url, params, fetch }) {
    let res = null;
    try {
      return { props: {
        users: [
          {
            "ID": 1,
            "CreatedAt": "2022-02-11T07:51:57.261657548Z",
            "firstname": "caca",
            "lastname": "caca",
            "email": "caca@caca.fr",
          },{
            "ID": 1,
            "CreatedAt": "2022-02-11T07:51:57.261657548Z",
            "firstname": "caca2",
            "lastname": "caca",
            "email": "caca@caca.fr",
          },{
            "ID": 1,
            "CreatedAt": "2022-02-11T07:51:57.261657548Z",
            "firstname": "caca3",
            "lastname": "caca",
            "email": "caca@caca.fr",
          },{
            "ID": 1,
            "CreatedAt": "2022-02-11T07:51:57.261657548Z",
            "firstname": "caca4",
            "lastname": "caca",
            "email": "caca@caca.fr",
          },{
            "ID": 1,
            "CreatedAt": "2022-02-11T07:51:57.261657548Z",
            "firstname": "caca5",
            "lastname": "caca",
            "email": "caca@caca.fr",
          }
        ],
      } };
      // res = await fetch(`${BACKEND_URL}/users`);
      // if (res.status > 400) {
      //   return { status: res.status, error: await res.text() };
      // }
      // return { props: { users: await res.json() } };
    } catch (err) {
      console.log(err);
      return { status: 500, error: new Error(`error fetching user: ${res}`) };
    }
  }
</script>

<script>
  import dayjs from 'dayjs';
  import * as yup from 'yup';

  import extractErrors from '$lib/extractErrors';
  import Pen from '../components/icon/Pen.svelte';
  import Trash from '../components/icon/Trash.svelte';
  import Modal from '../components/Modal.svelte';

  export let users;

  let userSelected = false;
  let values = {};
  let errors = {};
  const schema = yup.object?.().shape({
    firstname: yup.string().required('required'),
    lastname: yup.string().required('required'),
    email: yup.string().required('required').email('not an email'),
  });
  const onRefresh = async () => {
    try {
      users = await (await fetch(`${BACKEND_URL}/users`)).json();
    } catch (err) {
      // Pass
    }
  }
  const onSubmit = async () => {
    try {
      await schema.validate(values, { abortEarly: false });
    } catch (err) {
      console.log(JSON.stringify(err));
      errors = extractErrors(err);
      console.log(errors)
    }
    console.log(values);
    try {
      await fetch(`${BACKEND_URL}/users/${userSelected.ID}`, {
        method: 'POST',
        data: JSON.stringify(values),
      });
      await onRefresh()
    } catch (err) {
      // Pass
    }
  }
  const onRemove = async (user) => {
    console.log(user);
    try {
      await fetch(`${BACKEND_URL}/users/${user.ID}`, {
        method: 'DELETE',
      });
      await onRefresh()
    } catch (err) {
      // Pass
    }
  }
  $:console.log(userSelected);
  $:values = userSelected || {};
  $:errors = userSelected && {};
</script>

{#if userSelected}
  <Modal on:close={() => userSelected = false}>
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
    <div class="p-6 max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center">
      <p>{user.firstname} {user.lastname}</p>
      <p>{user.email}</p>
      <p>added {dayjs(user.CreatedAt).format('D/MM/YY')}</p>

      <div class="flex mt-2 h-5">
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-blue-500 hover:bg-blue-600"
          on:click|preventDefault|stopPropagation={() => userSelected = Object.assign({}, user)}
        >
          <Pen />
        </button>
        <button
          class=" rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 ml-2 transition bg-orange-500 hover:bg-orange-600"
          on:click|preventDefault|stopPropagation={() => onRemove(user)}
        >
          <Trash />
        </button>
      </div>
    </div>
  {/each}
</div>
