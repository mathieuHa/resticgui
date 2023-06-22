
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { page } from '$app/stores';
  import Toast from '$lib/Toast.svelte';
  import agent from '$utils/api';
  import { toasts } from '$utils/stores';
    import Modal from '$lib/Modal.svelte';

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
      toasts.addToast(err?.message, 'failure');
    }
  };

</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<Modal on:close={() => dispatch('close')}>
  <div class="modal-box">
    <h3 class="font-bold text-lg mb-4">Restoration snapshot #{snapshotId}</h3>
    <div class="form-control full-w">
      <label class="label" for="path">
        <span class="label-text">Where do you want restore files ?</span>
      </label>
      <label class="input-group" id="path">
        <span>Destination</span>
        <input type="text" placeholder="/home/restore" class="input input-bordered w-full" bind:value={path} />
      </label>
    </div>
    <Toast />
    <div class="modal-action">
      <button
        class="btn btn-info"
        on:click={() => dispatch('close')}
      >
        No
      </button>
      <button
        class="btn btn-error"
        on:click={restoreSnapshot}
      >
        Yes
      </button>
    </div>
  </div>
</Modal>