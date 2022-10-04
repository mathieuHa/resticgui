<script>
  import { createEventDispatcher } from 'svelte';
  import agent from '../utils/api';
  import { toasts } from '../utils/stores';

  import BackupPolicy from './Repository/BackupPolicy.svelte';
  import BackupSchedule from './Repository/BackupSchedule.svelte';
  import General from './Repository/General.svelte';
  import Modal from './Modal.svelte';
  import Provider from './Repository/Provider.svelte';
  import Paths from './Repository/FileSystem/Paths.svelte';

  export let repository = null;

  let step = 0;
  let values = null;
  let link = null;

  const dispatch = createEventDispatcher();
  
  const onCheck = async (id) => {
    if (!id) return;
    try {
      await agent.get(`/repositories/${id}/valid`);
      toasts.addToast('Repository valid');
    } catch (err) {
      toasts.addToast(err?.message || 'Repository not valid', 'failure');
    }
  }
  const onSubmit = async () => {
    console.log(values);
    try {
      console.log(link);
      await agent[repository?.id ? 'patch' : 'post'](link, values);
      toasts.addToast('Repository updated !');
      dispatch('refresh');
    } catch (err) {
      console.log(err);
      toasts.addToast(err, 'failure');
      console.log(err);
    }
  }

  $:if (repository === null) step = 0;
  $:values = Object.assign({}, repository) || step;
</script>

<Modal on:close={() => dispatch('close')} title="Repository">
  <form on:submit|preventDefault|stopPropagation={onSubmit}>
    <div
      title={!repository.id ? 'Give to repository a name' : null}
      class="flex justify-between flex-wrap {!repository.id ? 'cursor-not-allowed' : ''}"
    >
      {#each ['General', 'Paths', 'Schedule', 'Policy', 'Provider'] as stepped, i}
        <div
          class="flex-1 pb-1 px-4 rounded-t border border-gray-500 {!repository.id ? 'pointer-events-none' : ''} {step === i ? 'border-b-0 order-4 md:order-unset bg-white dark:bg-gray-800' : 'bg-gray-200 dark:bg-gray-700 cursor-pointer hover:bg-blue-500/[.5]'} {step === i ? 'bg-white-500/[.5]' : ''}"
          on:click={() => { step = i; }}
        >
          <span class="rounded-xl">{stepped}</span>
        </div>
      {/each}
    </div>
    <div class="flex flex-col p-4 border border-t-0 border-gray-500 rounded-b bg-white dark:bg-gray-800">
      {#if step === 0}
        <General
          initialValues={values}
          on:update={({ detail: general }) => {
            values = { ...repository, ...general };
            link = `/repositories/${repository.id || ''}`;
          }}
        />
      {:else if step === 1}
        <Paths
          initialPaths={values.edges && values.edges.paths
            ? values.edges.paths.map(({ relPath = '/root', included }) => included ? relPath : `!${relPath}`)
            : []}
          on:update={({ detail: paths }) => {
            values = paths;
            link = `/repositories/${repository.id}/paths`;
          }}
        />
      {:else if step === 2}
        <BackupSchedule
          initialBackupSchedule={(values.edges && values.edges.backupSchedule) || {}}
          on:update={({ detail: backupSchedule }) => {
            values = backupSchedule;
            link = `/repositories/${repository.id}/backupschedule`;
          }}
        />
      {:else if step === 3}
        <BackupPolicy
          initialBackupPolicy={(values.edges && values.edges.backupPolicy) || {}}
          on:update={({ detail: backupPolicy }) => {
            values = backupPolicy;
            link = `/repositories/${repository.id}/backuppolicy`;
          }}
        />
      {:else if step === 4}
        <Provider
          initialProvider={values}
          on:update={({ detail: providers }) => {
            values = providers.edges[providers.provider];
            link = `/repositories/${repository.id}/${providers.provider.toLowerCase()}`;
          }}
        />
      {/if}
    </div>
    <div class="flex mt-2">
      <div class="flex-1"></div>
      <button
        class="pb-1 px-5 mr-2 min-w-[100px] rounded transition bg-blue-500 hover:bg-blue-600 disabled:bg-gray-500"
        disabled={!repository.id}
        on:click|preventDefault={() => onCheck(repository.id)}
      >
        Check
      </button>
      <button type="submit" class="pb-1 px-5 min-w-[100px] rounded transition bg-blue-500 hover:bg-blue-600">
        Save
      </button>
    </div>
  </form>
</Modal>
