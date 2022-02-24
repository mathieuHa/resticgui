<script>
  import { createEventDispatcher } from 'svelte';

  export let initialProvider;

  const providers = {
    localProvider: ['repoName'],
    MinioS3Provider: ['accessKeyID', 'bucketName', 'port', 'secretAccessKey', 'scheme'],
    B2Provider: ['accountID', 'accountKey', 'bucketName', 'repoName'],
    AWSS3Provider: ['accessKeyID', 'secretAccessKey', 'bucketName', 'defaultRegion'],
  }

  let provider = null;
  let values = {};
  Object.keys(providers).forEach((backupProvider) => {
    values[backupProvider] = Object.assign({}, initialProvider[backupProvider]);
  });

  const dispatch = createEventDispatcher();

  $: console.log({ [provider]: values[provider] })
  $: dispatch('update', { [provider]: values[provider] });
</script>

<div class="flex">
  {#each ['localProvider', 'AWSS3Provider', 'MinioS3Provider', 'B2Provider'] as backupProvider}
    <img
      class="h-7 w-7 mr-2 cursor-pointer rounded {backupProvider === provider ? 'bg-yellow-500' : 'bg-transparent'}"
      src="/{backupProvider}.svg"
      on:click={() => { provider = backupProvider }}
    />
  {/each}
</div>
<div>
  {#if provider !== null}
    {#each providers[provider] as input}
      <div class="w-full flex justify-between">
        <label for="{input}">{ input }</label>
      </div>
      <input
        class="px-2 py-1 border-2 border-indigo-900 rounded dark:bg-gray-700"
        id="{input}"
        name="{input}"
        type="text"
        bind:value={values[provider][input]}
      />
    {/each}
  {/if}
</div>
