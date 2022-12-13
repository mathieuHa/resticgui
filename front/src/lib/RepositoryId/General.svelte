<script>
    export let repo;
    
    console.log(repo);
</script>
<div>
    <div class="grid grid-cols-2">
        <div class="bg-sky-500 rounded-full w-25 min-w-min text-center">Repository Name: {repo.name}</div>
        <div>Created Date: {repo.createdAt.split('.')[0].replace('T', ' ')}</div>
    </div>
    <br />
    
    <div class="grid grid-cols-4">
    {#if repo.active}
        <span class="bg-green-600 rounded-full w-24 min-w-min text-center">active</span>
    {:else}
        <div class="bg-red-600 rounded-full w-24 min-w-min text-center">active</div>
    {/if}
    {#if repo.initiated}
        <span class="bg-green-600 rounded-full w-24 min-w-min text-center">initiated</span>
    {:else}
        <div class="bg-red-600 rounded-full w-24 min-w-min text-center">initiated</div>
    {/if}
    {#if repo.valid}
        <span class="bg-green-600 rounded-full w-24 min-w-min text-center">valid</span>
    {:else}
        <div class="bg-red-600 rounded-full w-24 min-w-min text-center">valid</div>
    {/if}
    {#if repo.running}
        <span class="bg-green-600 rounded-full w-24 min-w-min text-center">running</span>
    {:else}
        <div class="bg-red-600 rounded-full w-24 min-w-min text-center">running</div>
    {/if}

    </div>
    <br />

    <div class="grid grid-cols-3">
        <div>Provider</div>
        <div>{repo.provider}</div>
        <div>
            <img
            class="h-10 w-10 mr-2 cursor-pointer rounded 'bg-black' transition-all"
            src="/{repo.provider}.svg"
            alt={repo.provider}
            />
        </div>
    </div>
    <br />

    <div class="grid grid-cols-2">
        <div>Schedule</div>
        {#if repo.edges.backupSchedule.choice === "interval"}
            <span class="bg-purple-300 rounded-full w-24 min-w-min text-center">interval {repo.edges.backupSchedule.interval}</span>
        {:else if repo.edges.backupSchedule.choice === "cron"}
            <div class="bg-purple-300 rounded-full w-24 min-w-min text-center">cron {repo.edges.backupSchedule.cron}</div>
        {:else}
            <div class="bg-purple-300 rounded-full w-24 min-w-min text-center">No planning set</div>
        {/if}
    </div>
    <br />

    {#if repo.active}
        <div class="grid grid-cols-2">
            <span class="bg-orange-600 rounded-full w-50 min-w-min text-center">{repo.edges.backupSchedule.nextRun}</span>
            <span class="bg-orange-600 rounded-full w-50 min-w-min text-center">{repo.edges.backupSchedule.lastRun}</span>
        </div>
        <br />
    {/if}

    <div class="grid grid-cols-2">
        <div>Paths</div>
        <div>
            {#each repo.edges.paths as path, i}
                "<span class={path.included === false ? 'text-orange-500' : 'text-green-500'}>{path.absPath}</span>"{i !== repo.edges.paths.length - 1 ? ', ' : ''}
            {/each}
        </div>
    </div>
    <br />

    <div class="grid grid-cols-3">
        <span class="bg-yellow-600 rounded-full w-50 min-w-min text-center">Repository Size {repo.repositorySize}</span>
        <span class="bg-yellow-600 rounded-full w-50 min-w-min text-center">Repository Total File Number {repo.repositoryFileCount}</span>
        <span class="bg-yellow-600 rounded-full w-50 min-w-min text-center">Repository Snapshot count {repo.snapshotCount}</span>
    </div>
    <br />
    
    <div>
        <div class="italic">Restic repository version: {repo.version} </div>
    </div>

</div>



