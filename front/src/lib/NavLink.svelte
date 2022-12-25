<script lang="ts">
	import { page } from '$app/stores';

	export let href: string;
	
  $: isActive = $page.url.pathname === $$props.href;
</script>

<a
	class="flex items-center h-9 rounded px-2 text-gray-800 transition-all hover:bg-yellow-200 dark:text-gray-200 dark:hover:bg-yellow-800"
	class:font-semibold={isActive}
	{href}
>
  <slot />
</a>
