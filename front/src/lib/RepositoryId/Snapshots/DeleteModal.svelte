<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';
    import Modal from '$lib/Modal.svelte';

  export let snapshotId: string;

  const dispatch = createEventDispatcher();

  const deleteSnapshot = async () => {
    const id = $page.params.id;
    try {
      await agent.delete(`/repositories/${id}/snapshots/${snapshotId}/forget`);
      toasts.addToast('Snapshot deleted');
      dispatch('delete');
      dispatch('close');
    } catch (err) {
      toasts.addToast(err?.message || 'Error while deleting snapshot', 'failure');
    }
  };

</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<Modal on:close={() => dispatch('close')}>
  <div class="modal-box" on:click|preventDefault|stopPropagation={() => {}}>
    <h3 class="font-bold text-lg">Warning</h3>
    <p class="py-4">Do you really want to delete this snapshot #{snapshotId} ?</p>
    <div class="modal-action">
      <button class="btn btn-info" on:click={() => dispatch('close')}>No</button>
      <button class="btn btn-error" on:click={() => deleteSnapshot()}>Yes</button>
    </div>
  </div>
</Modal>