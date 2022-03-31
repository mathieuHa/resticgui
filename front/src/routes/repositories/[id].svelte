<script>
  import { page } from '$app/stores';

  import agent from '../../utils/api';

  let repo = null;
  let snapshots = [];
  let tab = new URLSearchParams($page.url.search).get('tab') || null;

  const onRefresh = async (id) => {
    try {
      repo = await agent.get(`/repositories/${id}`);
      snapshots = await agent.get(`/repositories/${id}/snapshots`);
    } catch (err) {
      console.log(err)
    }
  }

  $: onRefresh($page.params.id);
</script>



{#if repo}
  {#if tab === 'snapshots'}
    {#each snapshots as snapshot}
      <pre class="text-sm">{JSON.stringify(snapshot, null, 1)}</pre>
    {/each}
  {:else}
    <pre class="text-sm">{JSON.stringify(repo, null, 1)}</pre>
  {/if}
{/if}
