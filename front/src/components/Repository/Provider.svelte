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


  $: console.log(values)
  $: dispatch('update', { provider, edges: { [provider]: values.edges[provider] } });
</script>

<div class="flex">
  {#each Object.keys(providers) as backupProvider}
    <img
      class="h-7 w-7 mr-2 cursor-pointer rounded {backupProvider === provider ? 'bg-yellow-500' : 'bg-transparent'}"
      src="/{backupProvider}.svg"
      on:click={() => { provider = backupProvider }}
    />
  {/each}
</div>
<p>{provider}</p>
{#if provider !== null}
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
