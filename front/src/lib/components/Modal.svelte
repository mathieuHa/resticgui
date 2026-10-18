<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    open = $bindable(false),
    title,
    wide = false,
    children,
    actions,
    onclose,
  }: {
    open?: boolean;
    title: string;
    wide?: boolean;
    children: Snippet;
    actions?: Snippet;
    onclose?: () => void;
  } = $props();

  let dialog: HTMLDialogElement | undefined = $state();

  $effect(() => {
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  });
</script>

<dialog
  bind:this={dialog}
  class="modal"
  onclose={() => {
    open = false;
    onclose?.();
  }}
>
  <div class="modal-box {wide ? 'max-w-4xl' : ''}">
    <h3 class="mb-4 text-lg font-bold">{title}</h3>
    {@render children()}
    {#if actions}
      <div class="modal-action">{@render actions()}</div>
    {/if}
  </div>
  <form method="dialog" class="modal-backdrop"><button aria-label="Close">close</button></form>
</dialog>
