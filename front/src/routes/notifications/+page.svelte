<script lang="ts">
  import { api, type Notifier, type NotifierInput } from '#lib/api/client.ts';
  import { attempt, notify } from '#lib/app.svelte.ts';
  import Modal from '#lib/components/Modal.svelte';
  import Bell from '@lucide/svelte/icons/bell';
  import Plus from '@lucide/svelte/icons/plus';

  let notifiers = $state<Notifier[] | null>(null);
  let editing = $state<(NotifierInput & { id?: number; service?: string }) | null>(null);
  let removing = $state<Notifier | null>(null);
  let error = $state('');
  let testing = $state<number | null>(null);

  const EXAMPLES = [
    ['ntfy', 'ntfy://ntfy.sh/my-backups'],
    ['Gotify', 'gotify://gotify.example.com/token'],
    ['Discord', 'discord://token@webhook-id'],
    ['Slack', 'slack://hook:token@webhook'],
    ['Telegram', 'telegram://token@telegram?chats=@channel'],
    ['Email', 'smtp://user:password@smtp.example.com:587/?from=backup@example.com&to=me@example.com'],
  ];

  async function load() {
    notifiers = await api.notifiers();
  }

  $effect(() => {
    load();
  });

  function add() {
    error = '';
    editing = { name: '', url: '', on_failure: true, on_success: false, on_missed: true, enabled: true };
  }

  function edit(n: Notifier) {
    error = '';
    editing = { id: n.id, service: n.service, name: n.name, url: '', on_failure: n.on_failure, on_success: n.on_success, on_missed: n.on_missed, enabled: n.enabled };
  }

  async function save() {
    if (!editing) return;
    const { id, service: _, ...input } = $state.snapshot(editing);
    input.name = input.name.trim();
    input.url = input.url?.trim();
    if (!input.name) error = 'Give the notifier a name';
    else if (!id && !input.url) error = 'Enter the notifier URL';
    else error = '';
    if (error) return;
    const done = await attempt(async () => {
      if (id) await api.updateNotifier(id, input);
      else await api.createNotifier(input);
    }, 'Could not save the notifier');
    if (done) {
      notify(`${input.name} saved`);
      editing = null;
      load();
    }
  }

  async function test(n: Notifier) {
    testing = n.id;
    if (await attempt(() => api.testNotifier(n.id), `Could not send to ${n.name}`)) notify(`Test sent to ${n.name}`);
    testing = null;
  }

  async function remove() {
    if (!removing) return;
    const n = removing;
    if (await attempt(() => api.deleteNotifier(n.id), 'Could not remove the notifier')) {
      notify(`${n.name} removed`);
      removing = null;
      load();
    }
  }

  function events(n: Notifier) {
    const on = [n.on_failure && 'failures', n.on_missed && 'missed backups', n.on_success && 'successes'].filter(Boolean);
    return on.length ? on.join(', ') : 'nothing';
  }
</script>

<svelte:head><title>Notifications · ResticGUI</title></svelte:head>

<div class="card bg-base-100 border-base-300 animate-rise mx-auto max-w-3xl border shadow-sm">
  <div class="card-body gap-4">
    <div class="flex items-center gap-2">
      <h1 class="card-title flex-1">Notifications</h1>
      <button class="btn btn-primary btn-sm" onclick={add}><Plus class="size-4" aria-hidden="true" />Add a notifier</button>
    </div>
    <p class="text-base-content/70 text-sm">
      Tell a phone or a chat when a backup, a cleanup or a check fails, when a scheduled backup is more than an hour late, or when backups succeed. Each
      notifier is a <a class="link" href="https://github.com/nicholas-fedor/shoutrrr" target="_blank" rel="noreferrer">Shoutrrr URL</a>.
    </p>
    {#if notifiers === null}
      <div class="flex flex-col gap-2" aria-busy="true">{#each [1, 2] as i (i)}<div class="skeleton h-12"></div>{/each}</div>
    {:else if notifiers.length === 0}
      <div class="flex flex-col items-center gap-2 py-6 text-center">
        <Bell class="text-base-content/40 size-8" aria-hidden="true" />
        <p class="text-base-content/70">No notifier yet. Add one to hear about failed backups.</p>
      </div>
    {:else}
      <ul class="flex flex-col gap-2" data-testid="notifiers">
        {#each notifiers as n (n.id)}
          <li class="bg-base-200 flex flex-col gap-2 rounded-lg p-3 sm:flex-row sm:items-center">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 font-medium">
                {n.name}
                <span class="badge badge-ghost badge-sm">{n.service}</span>
                {#if !n.enabled}<span class="badge badge-warning badge-sm">off</span>{/if}
              </div>
              <div class="text-base-content/70 text-xs">Sends {events(n)}</div>
            </div>
            <div class="flex gap-1">
              <button class="btn btn-sm" onclick={() => test(n)} disabled={testing === n.id}>{testing === n.id ? 'Sending…' : 'Send a test'}</button>
              <button class="btn btn-ghost btn-sm" onclick={() => edit(n)}>Edit</button>
              <button class="btn btn-ghost btn-sm" onclick={() => (removing = n)}>Remove</button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<Modal open={!!editing} title={editing?.id ? `Edit ${editing.name}` : 'Add a notifier'} onclose={() => (editing = null)}>
  {#if editing}
    <div class="flex flex-col gap-3">
      <label class="flex flex-col gap-1">
        <span class="label text-sm">Name</span>
        <input class="input input-sm w-full" placeholder="My phone" bind:value={editing.name} />
      </label>
      <label class="flex flex-col gap-1">
        <span class="label text-sm">{editing.id ? `URL, keep empty to keep the ${editing.service} URL saved` : 'URL'}</span>
        <input class="input input-sm w-full font-mono" placeholder="ntfy://ntfy.sh/my-backups" autocomplete="off" bind:value={editing.url} />
      </label>
      <details class="text-xs">
        <summary class="cursor-pointer">Examples</summary>
        <dl class="mt-2 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
          {#each EXAMPLES as [name, url] (name)}<dt>{name}</dt><dd class="font-mono break-all">{url}</dd>{/each}
        </dl>
      </details>
      <fieldset class="flex flex-col gap-1">
        <legend class="mb-1 text-sm">Send</legend>
        <label class="label cursor-pointer gap-2"><input type="checkbox" class="checkbox checkbox-sm" bind:checked={editing.on_failure} /><span class="text-base-content">Failures</span></label>
        <label class="label cursor-pointer gap-2"><input type="checkbox" class="checkbox checkbox-sm" bind:checked={editing.on_missed} /><span class="text-base-content">Missed scheduled backups</span></label>
        <label class="label cursor-pointer gap-2"><input type="checkbox" class="checkbox checkbox-sm" bind:checked={editing.on_success} /><span class="text-base-content">Successes too</span></label>
      </fieldset>
      <label class="label cursor-pointer gap-2"><input type="checkbox" class="toggle toggle-sm" bind:checked={editing.enabled} /><span class="text-base-content">On</span></label>
      {#if error}<p class="text-error text-sm">{error}</p>{/if}
    </div>
  {/if}
  {#snippet actions()}
    <button class="btn" onclick={() => (editing = null)}>Cancel</button>
    <button class="btn btn-primary" onclick={save}>Save</button>
  {/snippet}
</Modal>

<Modal open={!!removing} title="Remove {removing?.name}?">
  <p>ResticGUI stops sending to this notifier.</p>
  {#snippet actions()}
    <button class="btn" onclick={() => (removing = null)}>Cancel</button>
    <button class="btn btn-error" onclick={remove}>Remove</button>
  {/snippet}
</Modal>
