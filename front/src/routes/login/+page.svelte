<script lang="ts">
  import { ApiError, api } from '#lib/api/client.ts';

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

<form class="card bg-base-100 mx-auto mt-16 max-w-sm shadow" onsubmit={submit}>
  <div class="card-body gap-4">
    <h1 class="card-title">Log in</h1>
    <label class="floating-label">
      <span>Username</span>
      <input class="input w-full" placeholder="Username" autocomplete="username" bind:value={username} required />
    </label>
    <label class="floating-label">
      <span>Password</span>
      <input class="input w-full" type="password" placeholder="Password" autocomplete="current-password" bind:value={password} required />
    </label>
    {#if error}<p class="text-error text-sm">{error}</p>{/if}
    <button class="btn btn-primary" disabled={busy}>Log in</button>
  </div>
</form>
