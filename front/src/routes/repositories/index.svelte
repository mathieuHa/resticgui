<script>
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';

  import BackupPolicy from '../../components/Repository/BackupPolicy.svelte';
  import BackupSchedule from '../../components/Repository/BackupSchedule.svelte';
  import General from '../../components/Repository/General.svelte';
  import Provider from '../../components/Repository/Provider.svelte';

  import agent from '../../utils/api';
  import { toasts } from '../../utils/stores';
  import Eye from '../../components/Icon/Eye.svelte';
  import Pen from '../../components/Icon/Pen.svelte';
  import Trash from '../../components/Icon/Trash.svelte';
  import Modal from '../../components/Modal.svelte';
  import Paths from '../../components/Repository/FileSystem/Paths.svelte';

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
  const onCheck = async (id) => {
    if (!id) return;
    try {
      await agent.get(`/repositories/${id}/valid`);
      toasts.addToast('Repository valid');
    } catch (err) {
      toasts.addToast('Repository not valid', 'failure');
    }
  }
  const onSubmit = async () => {
    console.log(values);
    try {
      const data = await agent[values.id ? 'patch' : 'post'](`/repositories/${values.id || ''}`, values);
      toasts.addToast('Repository updated !');
      await onRefresh();
      initialValues = data;
    } catch (err) {
      toasts.addToast(err, 'failure');
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

  onMount(onRefresh);
</script>

{#if initialValues}
  <Modal on:close={() => { initialValues = null; }} title="Repository">
    <form on:submit|preventDefault|stopPropagation={onSubmit}>
      <div
        title={!initialValues.id ? 'Give to repository a name' : null}
        class="flex justify-between flex-wrap {!initialValues.id ? 'cursor-not-allowed' : ''}"
      >
        {#each ['General', 'Paths', 'Schedule', 'Policy', 'Provider'] as stepped, i}
          <div
            class="flex-1 text-center pb-1 px-2 rounded-t border border-gray-500 {!initialValues.id ? 'pointer-events-none' : ''} {step === i ? 'border-b-0 order-4 md:order-unset bg-white dark:bg-gray-800' : 'bg-gray-200 dark:bg-gray-700 cursor-pointer hover:bg-blue-500/[.5]'} {step === i ? 'bg-white-500/[.5]' : ''}"
            on:click={() => { step = i; }}
          >
            <span class="rounded-xl px-2">{stepped}</span>
          </div>
        {/each}
      </div>
      <div class="flex flex-col p-4 border border-t-0 border-gray-500 rounded-b bg-white dark:bg-gray-800">
        {#if step === 0}
          <General
            initialValues={values}
            on:update={({ detail: general }) => { values = { ...values, ...general }; }}
          />
        {:else if step === 1}
          <Paths
            initialPaths={values.edges && values.edges.paths
              ? values.edges.paths.map(({ relPath = '/root', included }) => included ? relPath : `!${relPath}`)
              : []}
            on:update={({ detail: paths }) => { values = { ...values, edges: { paths } }; }}
          />
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
      </div>
      <div class="flex mt-2">
        <div class="flex-1"></div>
        <button
          class="pb-1 px-5 mr-2 min-w-[100px] rounded transition bg-blue-500 hover:bg-blue-600 disabled:bg-gray-500"
          disabled={!initialValues.id}
          on:click|preventDefault={() => onCheck(initialValues.id)}
        >
          Check
        </button>
        <button type="submit" class="pb-1 px-5 min-w-[100px] rounded transition bg-blue-500 hover:bg-blue-600">
          Save
        </button>
      </div>
    </form>
  </Modal>
{/if}

<div class="w-full grid grid-cols-2 md:grid-cols-3 gap-3 mb-6">
  {#each repos as repo, i}
    <div class="p-3 flex flex-col max-w-sm bg-gray-200 dark:bg-gray-800 rounded shadow-lg">
      <p class="truncate text-lg font-bold {repo.active ? '' :'text-red-500'}">{repo.name}</p>
      <p class="truncate text-sm font-light my-2">{dayjs(repo.createdAt).format('DD/MM/YY')}</p>

      <div class="flex mt-auto pt-2 h-8 mb-1">
        <a href="/repositories/{repo.id}">
          <button class="rounded h-6 w-6 text-gray-900 dark:text-gray-100 p-1 transition bg-yellow-500 hover:bg-yellow-600">
            <Eye />
          </button>
        </a>
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
