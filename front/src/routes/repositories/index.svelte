<script>
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';
  import BackupPolicy from '../../components/BackupPolicy.svelte';
  import BackupSchedule from '../../components/BackupSchedule.svelte';
  import Provider from '../../components/Provider.svelte';

  import agent from '../../utils/api';
  import Eye from '../../components/Icon/Eye.svelte';
  import Pen from '../../components/Icon/Pen.svelte';
  import Trash from '../../components/Icon/Trash.svelte';
  import Modal from '../../components/Modal.svelte';
  import Paths from '../../components/FileSystem/Paths.svelte';

  let repos = [];
  let errors = {};
  let step = 0;
  let values = null;
  let initialValues = null;

  // const schema = yup.object?.().shape({
  //   name: yup.string().min(4, 'too short').required('required'),
  //   paths: yup.array(yup.object().shape({
  //     name: yup.string().min(1, 'too short').required('required'),
  //   })).min(1, 'too short').required('required'),
  //   backupSchedule: yup.object().shape({
  //     cron_syntax: yup.string().min(1, 'too short').required('required'),
  //   }),
  // });

  const onRefresh = async () => {
    try {
      repos = await agent.get(`/repositories`);
    } catch (err) {
      // Pass
    }
  }
  const onSubmit = async () => {
    console.log(values);
    try {
      const data = await agent[values.id ? 'patch' : 'post'](`/repositories/${values.id || ''}`, values);
      await onRefresh();
      initialValues = data;
    } catch (err) {
      console.log(err);
    }
  }
  const onRemove = async (repo) => {
    try {
      await agent.delete(`/repositories/${repo.id}`);
      await onRefresh()
    } catch (err) {
      console.log(err)
    }
  }

  $:if (initialValues === null) step = 0;
  $:values = Object.assign({}, initialValues) || step;
  $:console.log(initialValues, values)

  onMount(onRefresh);
</script>

{#if initialValues}
  <Modal on:close={() => { initialValues = null; }}>
    <div class="flex justify-between mb-2 flex-wrap">
      {#each ['name', 'paths', 'schedule', 'backup', 'provider'] as stepped, i}
        <div
          class="flex-1 px-2 border rounded-t-xl {step === i ? `order-4 md:order-unset shadow-sm shadow-yellow-500/[.5]` : ''} border-gray-500 {step === i ? 'bg-yellow-500/[.5]' : 'cursor-pointer hover:bg-orange-600/[.5]'}"
          on:click={() => { step = i; }}
        >
          <span class="rounded-xl px-2 ">{stepped}</span>
        </div>
      {/each}
    </div>
    <h1 class="mb-2">Update repository</h1>
    <form on:submit|preventDefault|stopPropagation={onSubmit} class="flex flex-col">
      {#if step === 0}
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
      {:else if step === 1}
        <div class="w-full flex justify-between">
          <label>Paths</label>
          <span class="text-red-500">{errors.path || ''}</span>
        </div>
        <p class="my-2">
          [
          {#each (values.edges && values.edges.paths || []) as path}
            "<span class={!path.included ? 'text-orange-500' : ''}>{path.relPath}</span>",&nbsp;
          {/each}
          ]
        </p>
        <div class="max-w-xl">
          <Paths
            initialPaths={values.edges && values.edges.paths
              ? values.edges.paths.map(({ relPath = '/root', included }) => included ? relPath : `!${relPath}`)
              : []}
            on:update={({ detail: paths }) => { values = { ...values, edges: { paths } }; }}
          />
        </div>
      {:else if step === 2}
        <BackupSchedule
          initialBackupSchedule={(values.edges && values.edges.backupSchedule) || {}}
          on:update={({ detail: backupSchedule }) => { values = { ...values, edges: { backupSchedule } }; }}
        />
      {:else if step === 3}
        <BackupPolicy
          initialBackupPolicy={(values.edges && values.edges.backupPolicy) || {}}
          on:update={({ detail: backupPolicy }) => { values = { ...values, edges: { backupPolicy } }; }}
        />
      {:else if step === 4}
        <Provider
          initialProvider={values}
          on:update={({ detail: providers }) => {
            values = {
              ...values,
              provider: providers.provider,
              edges: { ...providers.edges },
            };
          }}
        />
      {/if}
      <button
        class="mt-4 py-1 rounded transition bg-blue-500 hover:bg-blue-600"
        type="submit"
      >
        Save
      </button>
    </form>
  </Modal>
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each repos as repo, i}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p class="truncate text-lg font-bold">{repo.name}</p>
      <p class="truncate text-sm font-light my-2">{dayjs(repo.createdAt).format('DD/MM/YY')}</p>

      <div class="flex mt-auto pt-2 h-8 mb-1">
        <button class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-yellow-500 hover:bg-yellow-600">
          <a href="/repositories/{repo.id}">
            <Eye />
          </a>
        </button>
        <button
          class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 ml-2 transition bg-blue-500 hover:bg-blue-600"
          on:click|preventDefault|stopPropagation={() => { initialValues = Object.assign({}, repo); }}
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
    on:click={() => { initialValues = {}; }}
    class="text-3xl p-6 max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg items-center flex justify-center cursor-pointer"
  >
    +
  </div>
</div>
