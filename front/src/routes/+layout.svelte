<script lang="ts">
  import '../app.css';

  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { api, onUnauthorized, type Session } from '#lib/api/client.ts';
  import { attempt, jobs } from '#lib/app.svelte.ts';
  import Logo from '#lib/components/Logo.svelte';
  import ThemeToggle from '#lib/components/ThemeToggle.svelte';
  import Toasts from '#lib/components/Toasts.svelte';
  import LogOut from '@lucide/svelte/icons/log-out';
  import Plus from '@lucide/svelte/icons/plus';
  import WifiOff from '@lucide/svelte/icons/wifi-off';

  let { children } = $props();

  let session = $state<Session | null>(null);
  let failed = $state('');
  let version = $state('');

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
    api.version().then(
      (v) => (version = v.version),
      () => {},
    );
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

<div class="bg-base-200 flex min-h-screen flex-col">
  <nav class="bg-base-100/85 border-base-300 sticky top-0 z-30 border-b backdrop-blur">
    <div class="mx-auto flex max-w-6xl items-center gap-2 px-4 py-2">
      <a class="flex items-center gap-2 rounded-lg pr-2 text-lg font-semibold" href={resolve('/')}>
        <Logo />
        ResticGUI
      </a>
      <div class="flex-1"></div>
      {#if allowed && !jobs.connected}
        <span class="badge badge-warning badge-sm gap-1" title="Live updates are reconnecting"><WifiOff class="size-3" aria-hidden="true" />offline</span>
      {/if}
      <ThemeToggle />
      {#if allowed}
        <a class="btn btn-primary btn-sm" href={resolve('/repositories/new')} aria-label="New repository"><Plus class="size-4" aria-hidden="true" /><span class="hidden sm:inline">New repository</span></a>
        {#if session?.auth_enabled}
          <button class="btn btn-ghost btn-sm" onclick={logout} aria-label="Log out {session.username}" title="Log out {session.username}">
            <LogOut class="size-4" aria-hidden="true" /><span class="hidden sm:inline">{session.username}</span>
          </button>
        {/if}
      {/if}
    </div>
  </nav>

  <main class="mx-auto w-full max-w-6xl flex-1 p-4">
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

  {#if version}
    <footer class="text-base-content/70 p-4 text-center text-xs">ResticGUI {version}</footer>
  {/if}
</div>

<Toasts />
