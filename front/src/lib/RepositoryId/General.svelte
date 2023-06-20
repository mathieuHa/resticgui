<script lang="ts">
  import type { Repository } from '$utils/types';
  import Number from '$lib/Number.svelte';

  export let repo: Repository;
</script>

<div class="flex flex-wrap justify-between gap-2">
    <h1 class="text-xl font-bold leading-7 whitespace-nowrap">{repo.name}</h1>
    <p class="flex-1 text-sm font-light leading-7 text-right whitespace-nowrap">
      Created Date:
      {#if repo.createdAt?.split('.')[0]}{repo.createdAt?.split('.')[0]?.replace('T', ' ')}{/if}
    </p>
</div>
<br />

<div class="flex flex-wrap gap-2">
  <p>
    initiated:
    <span class="inline-block shadow rounded-full w-4 h-4 ml-2 mr-5 -mb-1 bg-{repo.initiated ? 'green' : 'red'}-500"/>
  </p>
  <p>
    valid:
    <span class="inline-block shadow rounded-full w-4 h-4 ml-2 mr-5 -mb-1 bg-{repo.valid ? 'green' : 'red'}-500"/>
  </p>
  <p>
    running:
    <span class="inline-block shadow rounded-full w-4 h-4 ml-2 mr-5 -mb-1 bg-{repo.running ? 'green' : 'red'}-500"/>
  </p>
</div>
<br />

<div>
  <img
    class="h-10 w-10 rounded bg-black transition-all"
    src="/ui/{repo.provider}.svg"
    alt={repo.provider}
  />
  <h2 class="text-md">
    <span class="font-bold mr-3">Provider:</span>
    {#if repo.provider}
      {repo.provider.replace('Provider', '')}
    {:else}
      no provider setup
    {/if}
  </h2>
</div>
<br />

<h2 class="text-md font-bold">Backup schedule</h2>
{#if repo.edges && repo.edges.backupSchedule}
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
    <span class="ml-3">{repo.edges.backupSchedule?.lastRun?.replace('+0000 UTC', '') || ''}</span>
  </p>
  <p class="flex-1">
    Next:
      {#if repo.active}
        {repo.edges.backupSchedule?.nextRun?.replace('+0000 UTC', '') || ''}
        <span class="inline-block shadow rounded-full w-4 h-4 ml-2 mr-5 -mb-1 bg-green-500"/>
      {:else}
        Backup disabled
        <span class="inline-block shadow rounded-full w-4 h-4 ml-2 mr-5 -mb-1 bg-red-500"/>
      {/if}
  </p>
  </div>
{:else}
  Schedule for backup is not setup
{/if}
  
<br />
<h2 class="text-md font-bold">Prune schedule</h2>
{#if repo.edges && repo.edges.pruneSchedule} 
  {#if repo.edges.pruneSchedule.choice === "interval"}
    <p>
      Interval:
      <span class="ml-3">{repo.edges.pruneSchedule.interval}</span>
    </p>
  {:else if repo.edges.pruneSchedule.choice === "cron"}
    <p>
      Cron:
      <span class="ml-3">{repo.edges.pruneSchedule.cronSyntax}</span>
    </p>
  {:else}
    <p>No planning set</p>
  {/if}
  <div class="flex flex-wrap gap-2 whitespace-nowrap">
    <p class="flex-1">
      Last:
      <span class="ml-3">
        {repo.edges.pruneSchedule?.lastRun?.replace('+0000 UTC', '') || ''}
      </span>
    </p>
    <p class="flex-1">
      Next:
      <span class="ml-3">
        {#if repo.pruneActive}
         {repo.edges.pruneSchedule?.nextRun?.replace('+0000 UTC', '') || ''}<span class="inline-block shadow rounded-full w-4 h-4 ml-2 mr-5 -mb-1 bg-green-500"/>
        {:else}
          Prune disabled <span class="inline-block shadow rounded-full w-4 h-4 ml-2 mr-5 -mb-1 bg-red-500"/>
        {/if}
      </span>
    </p>
  </div>
{:else}
  Schedule for prune is not setup
{/if}
<br />
<br />

<h2 class="text-md font-bold">Backup retention policy</h2>
{#if repo.edges && (repo.edges.backupPolicy===undefined || repo.edges.backupPolicy?.keepAll === true)}
  Keep all Backups
{:else if repo.edges && repo.edges.backupPolicy}
  <ul class="list-disc ml-5">
    <li>Keep the last {repo.edges.backupPolicy?.nbKeepLast} backups</li>
    {#if repo.edges.backupPolicy?.nbKeepHours && repo.edges.backupPolicy?.nbKeepHours !== 0}
      <li>Keep a backup per hour for {repo.edges.backupPolicy?.nbKeepHours} hour(s)</li>
    {/if}
    {#if repo.edges.backupPolicy?.nbKeepDays && repo.edges.backupPolicy?.nbKeepDays !== 0}
      <li>Keep a backup per day for {repo.edges.backupPolicy?.nbKeepDays} day(s)</li>
    {/if}
    {#if repo.edges.backupPolicy?.nbKeepWeeks && repo.edges.backupPolicy?.nbKeepWeeks !== 0}
      <li>Keep a backup per week for {repo.edges.backupPolicy?.nbKeepWeeks} week(s)</li>
    {/if}
    {#if repo.edges.backupPolicy?.nbKeepMonths && repo.edges.backupPolicy?.nbKeepMonths !== 0}
      <li>Keep a backup per month for {repo.edges.backupPolicy?.nbKeepMonths} month(s)</li>
    {/if}
  </ul>
{/if}
<br />
<br />

<h2 class="text-md font-bold">Paths</h2>
  {#if repo.edges?.paths && repo.edges?.paths.length > 0}
    <ul class="list-disc ml-5">
      {#each repo.edges?.paths as path, i}
        <li class="text-{!path.included ? 'orange' : 'green'}-500">{path.absPath}</li>
      {/each}
    </ul>
  {:else}
    No path to backup
  {/if}
  <br />
  <br />

<p>
  Size:
  <span class="ml-3"><Number number={repo.repositorySize || 0}/></span>
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


