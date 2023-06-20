<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Toast from '$lib/Toast.svelte';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';
  import { Provider, type Repository } from '$utils/types';

  import BackupPolicy from './BackupPolicy.svelte';
  import BackupSchedule from './BackupSchedule.svelte';
  import PruneSchedule from './PruneSchedule.svelte';
  import General from './General.svelte';
  import BackupProvider from './BackupProvider.svelte';
  import Paths from './FileSystem/Paths.svelte';
  import Modal from '$lib/Modal.svelte';

  export let repository: Repository;

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
        pruneActive: repo.pruneActive,
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
        nbKeepLast: repo.edges?.backupPolicy?.nbKeepLast || 0,
        nbKeepHours: repo.edges?.backupPolicy?.nbKeepHours || 0,
        nbKeepDays: repo.edges?.backupPolicy?.nbKeepDays || 0,
        nbKeepWeeks: repo.edges?.backupPolicy?.nbKeepWeeks || 0,
        nbKeepMonths: repo.edges?.backupPolicy?.nbKeepMonths || 0,
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
  let isCheckLoading = false;
  let isSubmitLoading = false;

  
  let values = stepper[0]?.defaultValues(repository);
  const onCheck = async (id?: number) => {
    isCheckLoading = true
    if (!id) return;
    try {
      await agent.get(`/${id}/valid`);
      toasts.addToast('Repository valid');
    } catch (err) {
      toasts.addToast(err?.message || 'Repository not valid', 'failure');
    }
    isCheckLoading = false
  }
  const onSubmit = async () => {
    isSubmitLoading = true
    let isPatch = !!repository?.id
    try {
      await agent[isPatch ? 'patch' : 'post'](isPatch ? link : '/repositories', values);
      toasts.addToast('Repository updated !');
      dispatch('refresh');
    } catch (err) {
      toasts.addToast(err?.message || 'Error while updating repository', 'failure');
      console.log(err);
    }
    isSubmitLoading = false
  }
  const onTabChange = (stepped: typeof stepper[0]) => {
    step = stepped.name;
    link = stepped.link;
    values = stepped.defaultValues(repository);
    toasts.clear();
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<Modal on:close={() => dispatch('close')}>
  <div class="modal-box" on:click|stopPropagation={() => {}}>
    <form class="max-h-full flex flex-col overflow-hidden" on:submit|preventDefault|stopPropagation={onSubmit}>
      <div
        title={!repository.id ? 'Give to repository a name' : null}
        class="tabs tabs-boxed rounded-t-xl rounded-b-none {!repository.id ? 'cursor-not-allowed' : ''}"
      >
        {#each stepper as stepped, i}
          <button
            class="tab {!repository.id ? 'pointer-events-none' : ''} {step === stepped.name ? 'tab-active' : ''}"
            on:click|preventDefault|stopPropagation={() => onTabChange(stepped)}
          >
            {stepped.name}
          </button>
        {/each}
      </div>
      <div class="overflow-y-auto relative flex-1 flex flex-col p-4 rounded-b-xl bg-base-200">
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
      <Toast />
      <div class="modal-action">
        <button
          class="btn btn-info {isCheckLoading ? 'loading' : ''}"
          disabled={!repository.id || isCheckLoading}
          on:click|preventDefault={() => onCheck(repository.id)}
        >
          Check
        </button>
        <button
          type="submit"
          class="btn btn-info {isSubmitLoading ? 'loading' : ''}"
          disabled={isSubmitLoading}
          on:click|preventDefault={() => onSubmit()}
        >
          Save
        </button>
      </div>
    </form>
  </div>
</Modal>
