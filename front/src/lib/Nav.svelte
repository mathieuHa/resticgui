<script lang="ts">
  import { page } from '$app/stores';
  import { refresh } from '$utils/stores';
	import NavLink from './NavLink.svelte';

	let isDark = false;
	if (typeof localStorage !== 'undefined') {
		if (
			localStorage.theme === 'dark' ||
			(!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)
		) {
			isDark = true;
		}
	}
	function toggleDarkMode(): void {
		if (isDark) {
			document.documentElement.classList.remove('dark');
			localStorage.theme = 'light';
			isDark = false;
		} else {
			document.documentElement.classList.add('dark');
			localStorage.theme = 'dark';
			isDark = true;
		}
	}
</script>

<nav class="relative mx-auto flex flex-wrap w-full max-w-2xl items-center justify-between text-gray-900 dark:text-gray-100 py-8 sm:pb-10">
	<ul class="md:flex md:order-unset md:w-auto">
		<li>
			<NavLink href="/">ResticGUI</NavLink>
		</li>
	</ul>
	<div class="flex items-center space-x-4">
    {#if $page.params.id}
      <button
        aria-label="Toggle Dark Mode"
        class="ml-1 flex h-9 w-9 items-center justify-center rounded ring-yellow-400 transition-all hover:bg-yellow-200 dark:text-gray-200 dark:hover:bg-yellow-800"
        on:click|preventDefault|stopPropagation={refresh.update}
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 491.236 491.236" width="16" height="16" fill="currentColor" stroke="currentColor">
          <path d="M55.89,262.818c-3-26-0.5-51.1,6.3-74.3c22.6-77.1,93.5-133.8,177.6-134.8v-50.4c0-2.8,3.5-4.3,5.8-2.6l103.7,76.2
            c1.7,1.3,1.7,3.9,0,5.1l-103.6,76.2c-2.4,1.7-5.8,0.2-5.8-2.6v-50.3c-55.3,0.9-102.5,35-122.8,83.2c-7.7,18.2-11.6,38.3-10.5,59.4
            c1.5,29,12.4,55.7,29.6,77.3c9.2,11.5,7,28.3-4.9,37c-11.3,8.3-27.1,6-35.8-5C74.19,330.618,59.99,298.218,55.89,262.818z
            M355.29,166.018c17.3,21.5,28.2,48.3,29.6,77.3c1.1,21.2-2.9,41.3-10.5,59.4c-20.3,48.2-67.5,82.4-122.8,83.2v-50.3
            c0-2.8-3.5-4.3-5.8-2.6l-103.7,76.2c-1.7,1.3-1.7,3.9,0,5.1l103.6,76.2c2.4,1.7,5.8,0.2,5.8-2.6v-50.4
            c84.1-0.9,155.1-57.6,177.6-134.8c6.8-23.2,9.2-48.3,6.3-74.3c-4-35.4-18.2-67.8-39.5-94.4c-8.8-11-24.5-13.3-35.8-5
            C348.29,137.718,346.09,154.518,355.29,166.018z"/>
        </svg>
      </button>
    {/if}
		<button
			aria-label="Toggle Dark Mode"
			class="ml-1 flex h-9 w-9 items-center justify-center rounded ring-yellow-400 transition-all hover:bg-yellow-200 dark:text-gray-200 dark:hover:bg-yellow-800"
			on:click|preventDefault|stopPropagation={toggleDarkMode}
		>
			{#if isDark}
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" class="h-5 w-5 text-gray-800 dark:text-yellow-100">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728	0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"/>
				</svg>
			{:else}
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" class="w-5 h-5 text-gray-800 dark:text-gray-200">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
				</svg>
			{/if}
		</button>
	</div>
</nav>
