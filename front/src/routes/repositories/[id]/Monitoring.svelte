<script lang="ts">
  import { untrack } from 'svelte';

  import { api, type Repository, type Settings } from '#lib/api/client.ts';
  import { attempt, notify } from '#lib/app.svelte.ts';
  import { draftsOf, EVENT_LABELS, EVENTS, hookProblems, newHook, toHookInputs, type HookEvent, type HookKind } from '#lib/hooks.ts';
  import Plus from '@lucide/svelte/icons/plus';
  import Trash from '@lucide/svelte/icons/trash-2';

  let { repo, onsaved }: { repo: Repository; onsaved: (r: Repository) => void } = $props();

  const initial = untrack(() => repo);
  let settings = $state<Settings>({ hook_commands: false, metrics: false });
  let healthcheck = $state('');
  let hooks = $state(draftsOf(initial.hooks));
  let errors = $state<string[]>([]);

  $effect(() => {
    api.settings().then((s) => (settings = s));
  });

  async function saveHealthcheck() {
    const url = healthcheck.trim();
    if (!/^https?:\/\/[^/\s]+/.test(url)) {
      errors = ['The Healthchecks URL starts with http:// or https://'];
      return;
    }
    errors = [];
    await attempt(async () => {
      onsaved(await api.setHealthcheck(repo.id, url));
      healthcheck = '';
      notify('Healthchecks URL saved');
    }, 'Could not save the Healthchecks URL');
  }

  async function removeHealthcheck() {
    await attempt(async () => {
      onsaved(await api.deleteHealthcheck(repo.id));
      notify('Healthchecks URL removed');
    }, 'Could not remove the Healthchecks URL');
  }

  function add(event: HookEvent, kind: HookKind) {
    hooks.push(newHook(event, kind));
  }

  async function saveHooks() {
    errors = hookProblems(hooks, settings.hook_commands);
    if (errors.length) return;
    await attempt(async () => {
      const saved = await api.setHooks(repo.id, toHookInputs($state.snapshot(hooks)));
      hooks = draftsOf(saved.hooks);
      onsaved(saved);
      notify('Hooks saved');
    }, 'Could not save the hooks');
  }
</script>

{#if errors.length}
  <ul class="alert alert-error flex-col items-start text-sm">
    {#each errors as e (e)}<li>{e}</li>{/each}
  </ul>
{/if}

<section class="card bg-base-100 border-base-300 border shadow-sm">
  <div class="card-body">
    <h2 class="card-title text-base">Healthchecks</h2>
    <p class="text-base-content/70 text-sm">
      ResticGUI pings this URL when a backup starts, succeeds or fails, so a service such as Healthchecks.io can tell you when backups stop.
    </p>
    {#if repo.healthcheck}
      <p class="text-sm">Pinging <span class="font-mono">{repo.healthcheck.host}</span>.</p>
    {/if}
    <label class="flex flex-col gap-1">
      <span class="label text-sm">{repo.healthcheck ? 'Replace the ping URL' : 'Ping URL'}</span>
      <input class="input input-sm w-full font-mono" type="url" placeholder="https://hc-ping.com/your-uuid" autocomplete="off" bind:value={healthcheck} />
    </label>
    <div class="card-actions justify-end">
      {#if repo.healthcheck}<button class="btn btn-ghost btn-sm" onclick={removeHealthcheck}>Stop pinging</button>{/if}
      <button class="btn btn-primary btn-sm" onclick={saveHealthcheck} disabled={!healthcheck.trim()}>Save ping URL</button>
    </div>
  </div>
</section>

<section class="card bg-base-100 border-base-300 border shadow-sm">
  <div class="card-body">
    <h2 class="card-title text-base">Hooks</h2>
    <p class="text-base-content/70 text-sm">
      Run a command or call a webhook around each backup. A failing hook before the backup stops it; the output of every hook goes into the job's log.
      {#if !settings.hook_commands}Commands are turned off on this server, set <code>hooks.commands: true</code> to allow them.{/if}
    </p>
    {#each EVENTS as event (event)}
      <div class="flex flex-col gap-2" data-testid="hooks-{event}">
        <div class="flex flex-wrap items-center gap-2">
          <h3 class="text-sm font-medium">{EVENT_LABELS[event]}</h3>
          <div class="flex-1"></div>
          <button class="btn btn-ghost btn-xs" onclick={() => add(event, 'command')} disabled={!settings.hook_commands}>
            <Plus class="size-3" aria-hidden="true" />Command
          </button>
          <button class="btn btn-ghost btn-xs" onclick={() => add(event, 'webhook')}><Plus class="size-3" aria-hidden="true" />Webhook</button>
        </div>
        {#each hooks as hook, i (hook)}
          {#if hook.event === event}
            <div class="bg-base-200 flex flex-col gap-2 rounded-lg p-2 sm:flex-row sm:items-end">
              <label class="flex min-w-0 flex-1 flex-col gap-1">
                <span class="label text-xs">{hook.kind === 'command' ? 'Command, run with sh -c' : hook.host ? `Webhook to ${hook.host}, enter a URL to replace it` : 'Webhook URL'}</span>
                {#if hook.kind === 'command'}
                  <input class="input input-sm w-full font-mono" placeholder="pg_dump app > /backup/app.sql" bind:value={hook.command} />
                {:else}
                  <input class="input input-sm w-full font-mono" type="url" placeholder="https://example.com/hook" autocomplete="off" bind:value={hook.url} />
                {/if}
              </label>
              <label class="flex flex-col gap-1 sm:w-32">
                <span class="label text-xs">Time limit (s)</span>
                <input class="input input-sm w-full" type="number" min="1" max="86400" bind:value={hook.timeout} />
              </label>
              <button class="btn btn-ghost btn-sm" aria-label="Remove hook {i + 1}" onclick={() => hooks.splice(i, 1)}><Trash class="size-4" aria-hidden="true" /></button>
            </div>
          {/if}
        {/each}
      </div>
    {/each}
    <div class="card-actions justify-end"><button class="btn btn-primary btn-sm" onclick={saveHooks}>Save hooks</button></div>
  </div>
</section>
