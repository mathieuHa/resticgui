<script>
  import { page } from '$app/stores';
  import { onMount } from 'svelte';

  import agent from '../../utils/api';

  let repo;

  const onRefresh = async (id) => {
    try {
      repo = await agent.get(`/repositories/${id}`);
      repo.backups = await agent.get(`/repositories/${id}/backups`);
    } catch (err) {
      console.log(err)
    }
  }

  $:onRefresh($page.params.id);

  onMount(onRefresh);
</script>

<pre class="text-sm">
  {JSON.stringify(repo, null, 1)}
</pre>
