<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';

  import BackupPolicy from './RepositoryModal/BackupPolicy.svelte';
  import BackupSchedule from './RepositoryModal/BackupSchedule.svelte';
  import PruneSchedule from './RepositoryModal/PruneSchedule.svelte';
  import General from './RepositoryModal/General.svelte';
  import Modal from './Modal.svelte';
  import Provider from './RepositoryModal/Provider.svelte';
  import Paths from './RepositoryModal/FileSystem/Paths.svelte';

  export let repository: any;

  const stepper = [
    { name: 'General', link:  `/repositories/${repository.id}` },
    { name: 'Paths', link:  `/repositories/${repository.id}/paths` },
    { name: 'Schedule', link:  `/repositories/${repository.id}/backupschedule` },
    { name: 'Policy', link:  `/repositories/${repository.id}/backuppolicy` },
    { name: 'Provider', link:  `/repositories/${repository.id}` },
    { name: 'Prune', link:  `/repositories/${repository.id}/pruneschedule` },
  ];
  let step = stepper[0].name;
  let values = null;
  let link: string = `/repositories/${repository.id}`;

  const dispatch = createEventDispatcher();
  
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
    console.log(link, values);
    try {
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
      {#each stepper as stepped, i}
        <button
          class="flex-1 pb-1 px-4 rounded-t border border-gray-500 {!repository.id ? 'pointer-events-none' : ''} {step === stepped.name ? 'border-b-0 order-4 md:order-unset bg-white dark:bg-gray-800' : 'bg-gray-200 dark:bg-gray-700 cursor-pointer hover:bg-blue-500/[.5]'} {step === stepped.name ? 'bg-white-500/[.5]' : ''}"
          on:click|preventDefault|stopPropagation={() => {
            step = stepped.name;
            link = stepped.link;
          }}
        >
          <span class="rounded-xl">{stepped.name}</span>
        </button>
      {/each}
    </div>
    <div class="flex flex-col p-4 border border-t-0 border-gray-500 rounded-b bg-white dark:bg-gray-800">
      {#if step === stepper[1].name}
        <Paths
          initialPaths={values.edges && values.edges.paths
            ? values.edges.paths.map(({ relPath = '/root', included }) => included ? relPath : `!${relPath}`)
            : []}
          on:update={({ detail: paths }) => (values = paths)}
        />
      {:else if step === stepper[2].name}
        <BackupSchedule
          initialBackupSchedule={(values.edges?.backupSchedule) || {}}
          on:update={({ detail: backupSchedule }) => {
            values = backupSchedule;
            link = `/repositories/${repository.id}/backupschedule`;
          }}
        />
      {:else if step === stepper[3].name}
        <BackupPolicy
          initialBackupPolicy={(values.edges?.backupPolicy) || {}}
          on:update={({ detail: backupPolicy }) => (values = backupPolicy)}
        />
      {:else if step === stepper[4].name}
        <Provider
          initialProvider={values}
          on:update={({ detail: providers }) => {
            values = providers.edges[providers.provider];
            link = `/repositories/${repository.id}/${providers.provider.toLowerCase()}`;
          }}
        />
      {:else if step === stepper[5].name}
        <PruneSchedule
          initialPruneSchedule={(values.edges?.pruneSchedule) || {}}
          on:update={({ detail: pruneSchedule }) => (values = pruneSchedule)}
        />
      {:else}
        <General
          initialValues={values}
          on:update={({ detail: general }) => (values = { ...repository, ...general })}
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
