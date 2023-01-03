
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';
  import Modal from "$lib/Modal.svelte";
  import Input from '$lib/Input.svelte';

  export let snapshotId: string;
  let path: string | null = null;

  const dispatch = createEventDispatcher();

  const restoreSnapshot = async () => {
    const id = $page.params.id;
    try {
      await agent.get(`/repositories/${id}/snapshots/${snapshotId}/restore?path=${path}`);
      toasts.addToast(`Backup restore of ${snapshotId} in ${path} has started`);
      dispatch('close');
    } catch (err) {
      toasts.addToast(err as string, 'failure');
    }
  };

</script>

<Modal title="Warning" on:close={() => dispatch('close')}>
  <div>
    Do You really want to restore this snapshot #{snapshotId}
    <Input
      secondary
      id="path"
      label="Path to restore to"
      type="text"
      bind:value={path}
    />
    <div class="flex gap-2 justify-end">
      <button
        class="bg-blue-500 hover:bg-blue-700 text-white py-1 px-4 rounded transition"
        on:click|preventDefault|stopPropagation={() => dispatch('close')}
      >
        No
      </button>
      <button
        class="bg-red-400 hover:bg-red-600 text-white py-1 px-4 rounded transition"
        on:click|preventDefault|stopPropagation={restoreSnapshot}
      >
        Yes
      </button>
    </div>
  </div>
</Modal>