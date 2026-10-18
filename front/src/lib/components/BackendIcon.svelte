<script lang="ts">
  import Cloud from '@lucide/svelte/icons/cloud';
  import HardDrive from '@lucide/svelte/icons/hard-drive';
  import Server from '@lucide/svelte/icons/server';

  let { backend, size = 'md' }: { backend: string; size?: 'md' | 'lg' } = $props();

  const icon = $derived(backend === 'local' ? HardDrive : ['s3', 'b2', 'azure', 'gs'].includes(backend) ? Cloud : Server);
  const Icon = $derived(icon);
</script>

<span class="bg-primary/10 text-primary grid shrink-0 place-items-center rounded-lg {size === 'lg' ? 'size-12' : 'size-10'}" aria-hidden="true">
  <Icon class={size === 'lg' ? 'size-6' : 'size-5'} />
</span>
