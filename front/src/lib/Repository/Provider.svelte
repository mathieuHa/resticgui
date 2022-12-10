<script>
  import { createEventDispatcher } from 'svelte';
  import Input from '../Input.svelte';

  export let initialProvider = {};

  const dispatch = createEventDispatcher();
  const providers = {
    localProvider: ['repoPath'],
    MinioS3Provider: ['accessKeyID', 'bucketName', 'port', 'secretAccessKey', 'scheme'],
    B2Provider: ['accountID', 'accountKey', 'bucketName', 'repoName'],
    AWSS3Provider: ['accessKeyID', 'secretAccessKey', 'bucketName', 'defaultRegion'],
  }

  let provider = initialProvider.provider || null;
  let values = { edges: {}, provider };
  Object.keys(providers).forEach((backupProvider) => {
    values.edges[backupProvider] = (initialProvider.edges && initialProvider.edges[backupProvider]) || {};
  });


  $: dispatch('update', { provider, edges: { [provider]: values.edges[provider] } });
</script>

<div class="flex justify-between">
  {#each Object.keys(providers) as backupProvider}
    <img
      class="h-10 w-10 mr-2 cursor-pointer rounded {backupProvider === provider ? 'bg-yellow-500' : 'bg-black'} transition-all"
      src="/{backupProvider}.svg"
      alt={backupProvider}
      on:click={() => { provider = backupProvider }}
    />
  {/each}
</div>
<p class="my-2">{provider.replace('provider')}</p>
{#if provider}
  {#each providers[provider] as input}
    <Input
      id={input}
      name={input}
      label={input}
      type="text"
      bind:value={values.edges[provider][input]}
    />
  {/each}
{/if}
