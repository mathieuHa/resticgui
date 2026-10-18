<script lang="ts">
  import type { Backend, Repository } from '$lib/api/client';
  import { BACKEND_LABELS, sortBackends, type RepositoryDraft } from '$lib/repository';

  import PathPicker from './PathPicker.svelte';

  // The storage settings of a repository: where it is and how to open it.
  let {
    draft = $bindable(),
    backends,
    existing,
  }: {
    draft: RepositoryDraft;
    backends: Backend[];
    /** the repository being edited, if any */
    existing?: Repository;
  } = $props();

  const backend = $derived(backends.find((b) => b.name === draft.backend));
  // a variable listed as both env and secret (access key IDs) is asked once, as a secret
  const envVars = $derived((backend?.env ?? []).filter((n) => !backend?.secrets.includes(n)));
  let picking = $state(false);
  let advanced = $state(false);
</script>

<div class="flex flex-col gap-4">
  <label class="floating-label">
    <span>Name</span>
    <input class="input w-full" placeholder="Name, such as Photos to NAS" bind:value={draft.name} required />
  </label>

  <fieldset class="fieldset">
    <legend class="fieldset-legend">Storage</legend>
    <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
      {#each sortBackends(backends) as b (b.name)}
        <label class="btn btn-sm justify-start {draft.backend === b.name ? 'btn-primary' : 'btn-outline'}">
          <input type="radio" class="hidden" name="backend" value={b.name} bind:group={draft.backend} />
          {BACKEND_LABELS[b.name] ?? b.name}
        </label>
      {/each}
    </div>
  </fieldset>

  <label class="floating-label">
    <span>Location</span>
    <input class="input w-full font-mono" placeholder={backend?.example || '/srv/restic-repo'} bind:value={draft.location} />
  </label>
  {#if draft.backend === 'local'}
    <div>
      <button type="button" class="btn btn-ghost btn-xs" onclick={() => (picking = !picking)}>
        {picking ? 'Hide folders' : 'Browse folders'}
      </button>
      {#if picking}
        <PathPicker
          onpick={(p) => {
            draft.location = p;
            picking = false;
          }}
        />
      {/if}
    </div>
  {:else if backend}
    <p class="-mt-2 text-xs opacity-60">For example <span class="font-mono">{backend.example}</span></p>
  {/if}

  <label class="floating-label">
    <span>Repository password</span>
    <input
      type="password"
      class="input w-full"
      autocomplete="new-password"
      placeholder={existing ? 'Repository password (leave empty to keep it)' : 'Repository password'}
      bind:value={draft.password}
    />
  </label>
  {#if !existing}
    <p class="-mt-2 text-xs opacity-60">restic encrypts the repository with it. Without it the backups cannot be read, so keep a copy.</p>
  {/if}

  {#if envVars.length || backend?.secrets.length}
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Credentials and settings</legend>
      {#each envVars as name (name)}
        <label class="input input-sm w-full">
          <span class="label w-56 font-mono text-xs">{name}{backend?.required.includes(name) ? ' *' : ''}</span>
          <input bind:value={draft.env[name]} />
        </label>
      {/each}
      {#if existing && !draft.replaceSecrets}
        <div class="flex items-center gap-2 text-sm">
          <span class="opacity-70">
            {existing.secrets.length ? `Stored: ${existing.secrets.join(', ')}` : 'No credentials stored'}
          </span>
          <button type="button" class="btn btn-xs" onclick={() => (draft.replaceSecrets = true)}>Replace</button>
        </div>
      {:else}
        {#each backend?.secrets ?? [] as name (name)}
          <label class="input input-sm w-full">
            <span class="label w-56 font-mono text-xs">{name}{backend?.required.includes(name) ? ' *' : ''}</span>
            <input type="password" autocomplete="off" bind:value={draft.secrets[name]} />
          </label>
        {/each}
        {#if existing}
          <p class="text-xs opacity-60">Saving replaces all the stored credentials with these.</p>
        {/if}
      {/if}
    </fieldset>
  {/if}

  <div class="collapse-arrow bg-base-200 collapse {advanced ? 'collapse-open' : ''}">
    <button type="button" class="collapse-title text-left text-sm font-medium" onclick={() => (advanced = !advanced)}>Advanced</button>
    <div class="collapse-content flex flex-col gap-3">
      <label class="select select-sm w-full">
        <span class="label">Compression</span>
        <select bind:value={draft.compression}>
          <option value="auto">auto</option>
          <option value="max">max</option>
          <option value="off">off</option>
        </select>
      </label>
      <label class="input input-sm w-full">
        <span class="label">Skip files larger than</span>
        <input placeholder="e.g. 2G" bind:value={draft.exclude_larger_than} />
      </label>
      <label class="input input-sm w-full">
        <span class="label">Host name in snapshots</span>
        <input placeholder="this machine's name" bind:value={draft.host} />
      </label>
      <label class="input input-sm w-full">
        <span class="label">CA certificate file</span>
        <input placeholder="/etc/ssl/my-ca.pem" bind:value={draft.cacert} />
      </label>
      <label class="flex flex-col gap-1 text-sm">
        <span>restic options, one <span class="font-mono">key=value</span> per line</span>
        <textarea class="textarea textarea-sm w-full font-mono" rows="2" placeholder="s3.storage-class=STANDARD_IA" bind:value={draft.extended}></textarea>
      </label>
    </div>
  </div>
</div>
