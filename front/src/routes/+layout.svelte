<script lang="ts">
  import '../app.css';

  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { api, onUnauthorized, type Session } from '#lib/api/client.ts';
  import { attempt, jobs } from '#lib/app.svelte.ts';
  import Toasts from '#lib/components/Toasts.svelte';

  let { children } = $props();

  let session = $state<Session | null>(null);
  let failed = $state('');

  const onLogin = $derived(page.route.id === '/login');

  async function check() {
    try {
      session = await api.session();
      failed = '';
    } catch (err) {
      failed = err instanceof Error ? err.message : String(err);
    }
  }

  onUnauthorized(() => {
    if (session) session.authenticated = false;
  });

  $effect(() => {
    check();
  });

  const allowed = $derived(!!session && (!session.auth_enabled || session.authenticated));

  $effect(() => {
    if (!session) return;
    if (!allowed && !onLogin) {
      goto(resolve('/login'), { replaceState: true });
    } else if (allowed && onLogin) {
      goto(resolve('/'), { replaceState: true });
    }
  });

  $effect(() => {
    if (!allowed) return;
    jobs.connect(api.eventsURL());
    return () => jobs.disconnect();
  });

  async function logout() {
    if (await attempt(() => api.logout())) {
      jobs.disconnect();
      await check();
    }
  }
</script>

<div class="bg-base-200 min-h-screen">
  <nav class="navbar bg-base-100 shadow-sm">
    <div class="flex-1">
      <a class="btn btn-ghost text-xl" href={resolve('/')}>ResticGUI</a>
    </div>
    {#if allowed}
      <div class="flex items-center gap-2">
        {#if !jobs.connected}
          <span class="badge badge-warning badge-sm" title="Live updates are reconnecting">offline</span>
        {/if}
        <a class="btn btn-primary btn-sm" href={resolve('/repositories/new')}>New repository</a>
        {#if session?.auth_enabled}
          <button class="btn btn-ghost btn-sm" onclick={logout}>Log out {session.username}</button>
        {/if}
      </div>
    {/if}
  </nav>

  <main class="mx-auto max-w-6xl p-4">
    {#if failed}
      <div class="alert alert-error">
        <span>Cannot reach the ResticGUI server: {failed}</span>
        <button class="btn btn-sm" onclick={check}>Retry</button>
      </div>
    {:else if !session}
      <span class="loading loading-spinner"></span>
    {:else if allowed || onLogin}
      {@render children()}
    {/if}
  </main>
</div>

<Toasts />
