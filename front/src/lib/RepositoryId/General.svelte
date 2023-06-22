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

<div class="stats auto-cols-fr shadow-xl w-full font-bold">
  <div class="stat">
    <p class="sm:text-xl {repo.initiated ? '' : 'text-red-500'}">initiated</p>
  </div>
  <div class="stat">
    <p class="sm:text-xl {repo.valid ? '' : 'text-red-500'}">valid</p>
  </div>
  <div class="stat">
    <p class="sm:text-xl {repo.running ? '' : 'text-red-500'}">running</p>
  </div>
</div>
<br />
<br />

<div>
  <div class="avatar rounded">
    <div class="h-10 w-10 rounded bg-black">
      <img src="/ui/{repo.provider}.svg" alt={repo.provider} />
    </div>
  </div>
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

<div class="stats stats-vertical sm:stats-horizontal auto-cols-fr shadow-xl w-full">
  <div class="stat">
    {#if repo.edges && repo.edges.backupSchedule}
      {#if repo.edges.backupSchedule.choice === "interval"}
          <p class="stat-title">Backup Interval</p>
          <p class="stat-value {repo.pruneActive ? '' : 'text-red-500'}">{repo.edges.backupSchedule.interval}</p>
      {:else if repo.edges.backupSchedule.choice === "cron"}
        <p class="stat-title">Backup Cron</p>
        <p class="stat-value {repo.pruneActive ? '' : 'text-red-500'}">{repo.edges.backupSchedule.cronSyntax}</p>
      {:else}
        <p>--</p>
      {/if}
      <p class="stat-desc">Last: {repo.edges.backupSchedule?.lastRun?.replace('+0000 UTC', '') || ''}</p>
      <p class="stat-desc">Next: {repo.edges.backupSchedule?.nextRun?.replace('+0000 UTC', '') || ''}</p>
    {:else}
    <p class="stat-title">Backup schedule not setup</p>
    {/if}
  </div>
  <div class="stat">
    {#if repo.edges && repo.edges.pruneSchedule}
      {#if repo.edges.pruneSchedule.choice === "interval"}
          <p class="stat-title">Prune Interval</p>
          <p class="stat-value {repo.pruneActive ? '' : 'text-red-500'}">{repo.edges.pruneSchedule.interval}</p>
      {:else if repo.edges.pruneSchedule.choice === "cron"}
        <p class="stat-title">Prune Cron</p>
        <p class="stat-value {repo.pruneActive ? '' : 'text-red-500'}">{repo.edges.pruneSchedule.cronSyntax}</p>
      {:else}
        <p>--</p>
      {/if}
      <p class="stat-desc">Last: {repo.edges.pruneSchedule?.lastRun?.replace('+0000 UTC', '') || ''}</p>
      <p class="stat-desc">Next: {repo.edges.pruneSchedule?.nextRun?.replace('+0000 UTC', '') || ''}</p>
    {:else}
      <p class="stat-title">Prune schedule not setup</p>
    {/if}
  </div>
</div>
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


