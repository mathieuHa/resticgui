<script lang="ts">
	import { page } from '$app/stores';

	export let href: string;
	
  $: isActive = $page.url.pathname === $$props.href;
</script>

<a
	class="rounded-lg p-1 text-gray-800 transition-all hover:bg-yellow-200 dark:text-gray-200 dark:hover:bg-yellow-800 sm:px-3 sm:py-2 inline-block"
	class:font-semibold={isActive}
	{href}
	><span class="capsize"><slot /> </span>
</a>
