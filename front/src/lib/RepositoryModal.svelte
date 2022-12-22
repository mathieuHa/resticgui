<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';

  import BackupPolicy from './RepositoryModal/BackupPolicy.svelte';
  import BackupSchedule from './RepositoryModal/BackupSchedule.svelte';
  import PruneSchedule from './RepositoryModal/PruneSchedule.svelte';
  import General from './RepositoryModal/General.svelte';
  import Modal from './Modal.svelte';
  import BackupProvider from './RepositoryModal/BackupProvider.svelte';
  import Paths from './RepositoryModal/FileSystem/Paths.svelte';
  import { Provider, type Repository } from '$utils/types';

  export let repository: any;

  const dispatch = createEventDispatcher();
  const stepper = [
    {
      name: 'General',
      link:  `/repositories/${repository.id}`,
      component: General,
      defaultValues: (repo: Repository) => ({
        id: repo.id,
        name: repo.name,
        password: repo.password,
        active: repo.active,
      }),
    },
    {
      name: 'Paths',
      link:  `/repositories/${repository.id}/paths`,
      component: Paths,
      defaultValues: (repo: Repository) => repo.edges?.paths || [],
    },
    {
      name: 'Schedule',
      link:  `/repositories/${repository.id}/backupschedule`,
      component: BackupSchedule,
      defaultValues: (repo: Repository) => ({
        choice: repo.edges?.backupSchedule?.choice,
        cronSyntax: repo.edges?.backupSchedule?.cronSyntax,
        interval: repo.edges?.backupSchedule?.interval,
      }),
    },
    {
      name: 'Policy',
      link:  `/repositories/${repository.id}/backuppolicy`,
      component: BackupPolicy,
      defaultValues: (repo: Repository) => ({
        keepAll: repo.edges?.backupPolicy?.keepAll,
        nbLastBackup: repo.edges?.backupPolicy?.nbLastBackup || 0,
        oneByHour: repo.edges?.backupPolicy?.oneByHour || 0,
        oneByDay: repo.edges?.backupPolicy?.oneByDay || 0,
        oneByMonth: repo.edges?.backupPolicy?.oneByMonth || 0,
      }),
    },
    {
      name: 'Provider',
      link:  `/repositories/${repository.id}/provider`,
      component: BackupProvider,
      defaultValues: (repo: Repository) => ({
        provider: repo.provider,
        [Provider.LocalProvider]: repo.edges?.[Provider.LocalProvider],
        [Provider.MinioS3Provider]: repo.edges?.[Provider.MinioS3Provider],
        [Provider.B2Provider]: repo.edges?.[Provider.B2Provider],
        [Provider.AWSS3Provider]: repo.edges?.[Provider.AWSS3Provider],
      }),
    },
    {
      name: 'Prune',
      link:  `/repositories/${repository.id}/pruneschedule`,
      component: PruneSchedule,
      defaultValues: (repo: Repository) => ({
        choice: repo.edges?.pruneSchedule?.choice,
        cronSyntax: repo.edges?.pruneSchedule?.cronSyntax,
        interval: repo.edges?.pruneSchedule?.interval,
      }),
    },
  ];
  let step = stepper[0]?.name;
  let link = `/repositories/${repository.id}`;

  
  let values = stepper[0]?.defaultValues(repository);
  const onCheck = async (id: number) => {
    if (!id) return;
    try {
      await agent.get(`/${id}/valid`);
      toasts.addToast('Repository valid');
    } catch (err) {
      toasts.addToast(err?.message || 'Repository not valid', 'failure');
    }
  }
  const onSubmit = async () => {
    let isPatch = !!repository?.id
    try {
      await agent[isPatch ? 'patch' : 'post'](isPatch ? link : '/repositories', values);
      toasts.addToast('Repository updated !');
      dispatch('refresh');
    } catch (err) {
      toasts.addToast(err, 'failure');
      console.log(err);
    }
  }
</script>

<Modal on:close={() => dispatch('close')} title="Repository">
  <form class="max-h-full flex flex-col overflow-hidden" on:submit|preventDefault|stopPropagation={onSubmit}>
    <div
      title={!repository.id ? 'Give to repository a name' : null}
      class="flex justify-between flex-wrap {!repository.id ? 'cursor-not-allowed' : ''}"
    >
      {#each stepper as stepped, i}
        <button
          class="flex-1 pb-1 px-4 rounded-t border border-gray-500 {!repository.id ? 'pointer-events-none' : ''} {step === stepped.name ? 'border-b-0 order-4 md:order-unset bg-white dark:bg-gray-800' : 'bg-gray-200 dark:bg-gray-700 cursor-pointer hover:bg-blue-500/[.5]'} {step === stepped.name ? 'bg-white-500/[.5]' : ''}"
          on:click|preventDefault|stopPropagation={() => {
            step = stepped.name;
            link = stepped.link;
            values = stepped.defaultValues(repository);
          }}
        >
          <span class="rounded-xl">{stepped.name}</span>
        </button>
      {/each}
    </div>
    <div class="overflow-y-auto relative flex-1 flex flex-col p-4 border border-t-0 border-gray-500 rounded-b bg-white dark:bg-gray-800">
      {#each stepper as stepped}
        {#if step === stepped.name}
          <svelte:component
            this={stepped.component}
            initialValues={stepped.defaultValues(repository)}
            on:update={({ detail }) => (values = detail)}
          />
        {/if}
      {/each}
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
