<script>
  import { createEventDispatcher } from 'svelte';

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
{#if provider !== null}
  <div>
    {#each providers[provider] as input}
      <div class="w-full flex justify-between">
        <label for={input}>{input}</label>
      </div>
      <input
        class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
        id={input}
        name={input}
        type="text"
        bind:value={values.edges[provider][input]}
      />
    {/each}
  </div>
{/if}
