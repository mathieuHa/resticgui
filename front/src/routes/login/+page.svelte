<script lang="ts">
  import { ApiError, api } from '#lib/api/client.ts';
  import Logo from '#lib/components/Logo.svelte';

  let username = $state('');
  let password = $state('');
  let error = $state('');
  let busy = $state(false);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    busy = true;
    error = '';
    try {
      await api.login(username, password);
      // the layout sees the new session and leaves this page
      location.reload();
    } catch (err) {
      error = err instanceof ApiError ? err.message : String(err);
      busy = false;
    }
  }
</script>

<svelte:head><title>Log in · ResticGUI</title></svelte:head>

<form class="card bg-base-100 border-base-300 animate-rise mx-auto mt-12 max-w-sm border shadow-sm" onsubmit={submit}>
  <div class="card-body gap-4">
    <div class="flex items-center gap-3">
      <Logo />
      <div>
        <h1 class="text-lg font-semibold">Log in</h1>
        <p class="text-base-content/70 text-sm">to your ResticGUI</p>
      </div>
    </div>
    <label class="floating-label">
      <span>Username</span>
      <input class="input w-full" placeholder="Username" autocomplete="username" bind:value={username} required />
    </label>
    <label class="floating-label">
      <span>Password</span>
      <input class="input w-full" type="password" placeholder="Password" autocomplete="current-password" bind:value={password} required />
    </label>
    {#if error}<p class="text-error text-sm" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" disabled={busy}>
      {#if busy}<span class="loading loading-spinner loading-sm"></span>{/if}
      Log in
    </button>
  </div>
</form>
