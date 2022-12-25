<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Input from '$lib/Input.svelte';
  import { Provider, type AWSS3Provider, type B2Provider, type LocalProvider, type MinioS3Provider } from '$utils/types';

  export let initialValues: {
    provider: Provider
    [Provider.LocalProvider]: LocalProvider
    [Provider.MinioS3Provider]: MinioS3Provider
    [Provider.B2Provider]: B2Provider
    [Provider.AWSS3Provider]: AWSS3Provider
  };

  const providers: Record<Provider, string[]> = {
    [Provider.LocalProvider]: ['repoPath'],
    [Provider.MinioS3Provider]: ['accessKeyID', 'bucketName', 'port', 'secretAccessKey', 'scheme'],
    [Provider.B2Provider]: ['accountID', 'accountKey', 'bucketName', 'repoName'],
    [Provider.AWSS3Provider]: ['accessKeyID', 'secretAccessKey', 'bucketName', 'defaultRegion'],
  }

  let provider = initialValues.provider || 'localProvider';
  let values = JSON.parse(JSON.stringify(initialValues));
  Object.keys(providers).forEach((backupProvider) => {
    values[backupProvider] = initialValues[backupProvider as Provider] || {};
  });

  const dispatch = createEventDispatcher();

  $: dispatch('update', { provider, [provider]: values[provider] });
</script>

<div class="flex justify-between">
  {#each Object.keys(providers) as backupProvider}
    <button
      class=" mr-2"
      on:click|preventDefault|stopPropagation={() => (provider = backupProvider)}
    >
      <img
        class="h-10 w-10 cursor-pointer rounded {backupProvider === provider ? 'bg-yellow-500' : 'bg-black'} transition-all"
        src="/{backupProvider}.svg"
        alt={backupProvider}
        
      />
    </button>
   
  {/each}
</div>
<p class="my-2">{provider.replace('provider', '')}</p>
{#if provider}
  {#each providers[provider] ?? [] as input}
    <Input
      id={input}
      name={input}
      label={input}
      type="text"
      bind:value={values[provider][input]}
    />
  {/each}
{/if}
