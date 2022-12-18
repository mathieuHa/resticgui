<script>
    export let repo;
</script>

<div class="flex flex-wrap justify-between gap-2 mt-5">
    <h1 class="text-xl font-bold leading-7 whitespace-nowrap">{repo.name}</h1>
    <p class="flex-1 text-sm font-light leading-7 text-right whitespace-nowrap">
      Created Date: {repo.createdAt.split('.')[0].replace('T', ' ')}
    </p>
</div>
<br />

<div class="flex flex-wrap gap-2">
  <p class="flex items-center">
    active:
    <span class="ml-2 block bg-{repo.active ? 'green' : 'red'}-500 shadow rounded-full w-4 h-4 mr-5 mt-1"/>
  </p>
  <p class="flex items-center">
    initiated:
    <span class="ml-2 block bg-{repo.initiated ? 'green' : 'red'}-500 shadow rounded-full w-4 h-4 mr-5 mt-1"/>
  </p>
  <p class="flex items-center">
    valid:
    <span class="ml-2 block bg-{repo.valid ? 'green' : 'red'}-500 shadow rounded-full w-4 h-4 mr-5 mt-1"/>
  </p>
  <p class="flex items-center">
    running:
    <span class="ml-2 block bg-{repo.running ? 'green' : 'red'}-500 shadow rounded-full w-4 h-4 mr-5 mt-1"/>
  </p>
</div>
<br />

<div>
  <img
    class="h-10 w-10 rounded bg-black transition-all"
    src="/{repo.provider}.svg"
    alt={repo.provider}
  />
  <h2 class="text-md">
    <span class="font-bold mr-3">Provider:</span>
    {repo.provider.replace('Provider', '')}
  </h2>
</div>
<br />

{#if repo.active}
  <h2 class="text-md font-bold">Backup schedule</h2>
  {#if repo.edges.backupSchedule.choice === "interval"}
    <p>
      Interval:
      <span class="ml-3">{repo.edges.backupSchedule.interval}</span>
    </p>
  {:else if repo.edges.backupSchedule.choice === "cron"}
    <p>
      Cron:
      <span class="ml-3">{repo.edges.backupSchedule.cronSyntax}</span>
    </p>
  {:else}
    <p>No planning set</p>
  {/if}
  <div class="flex flex-wrap gap-2 whitespace-nowrap">
    <p class="flex-1">
      Last:
      <span class="ml-3">
        {repo.edges.backupSchedule?.lastRun?.replace('+0000 UTC', '') || ''}
      </span>
    </p>
    <p class="flex-1">
      Next:
      <span class="ml-3">
        {repo.edges.backupSchedule?.nextRun?.replace('+0000 UTC', '') || ''}
      </span>
    </p>
  </div>
  <br />
{/if}

<h2 class="text-md font-bold">Paths</h2>
<ul class="list-disc ml-5">
  {#if repo.edges?.paths && repo.edges?.paths.length > 0}
    {#each repo.edges?.paths as path, i}
      <li class="text-{!path.included ? 'orange' : 'green'}-500">{path.absPath}</li>
    {/each}
  {/if}
</ul>
<br />

<p>
  Size:
  <span class="ml-3">{repo.repositorySize || 0}</span>
</p>
<p>
  Total file number:
  <span class="ml-3">{repo.repositoryFileCount || 0}</span>
</p>
<p>
  Snapshots count:
  <span class="ml-3">{repo.snapshotCount || 0}</span>
</p>
<br />
<p class="text-sm font-light italic">
  Restic repository version:
  <span class="font-normal ml-3 not-italic">{repo.version}</span>
</p>


