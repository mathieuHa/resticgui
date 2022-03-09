<script context="module">
	import {
		SITE_URL, SITE_TITLE, SITE_DESCRIPTION, DEFAULT_OG_IMAGE,
	} from '../utils/siteConfig';
	export const prerender = true; // index page is most visited, lets prerender
</script>

<svelte:head>
	<title>{SITE_TITLE}</title>
	<link rel="canonical" href={SITE_URL} />
	<link rel="alternate" type="application/rss+xml" href={SITE_URL + '/api/rss.xml'} />
	<meta property="og:url" content={SITE_URL} />
	<meta property="og:type" content="article" />
	<meta property="og:title" content={SITE_TITLE} />
	<meta name="Description" content={SITE_DESCRIPTION} />
	<meta property="og:description" content={SITE_DESCRIPTION} />
	<meta property="og:image" content={DEFAULT_OG_IMAGE} />
	<meta name="twitter:card" content="summary" />
	<meta name="twitter:title" content={SITE_TITLE} />
	<meta name="twitter:description" content={SITE_DESCRIPTION} />
	<meta name="twitter:image" content={DEFAULT_OG_IMAGE} />
</svelte:head>

<div class="flex flex-col-reverse items-start sm:flex-row w-full">
	<div class="flex flex-col w-full">
		<h1 class="mb-3 text-3xl font-bold tracking-tight text-black dark:text-white md:text-5xl">
			This is
			<span class="relative ml-2 inline-block before:absolute before:-inset-1 before:block before:-skew-y-3 before:bg-red-500">
				<span class="relative skew-y-3 text-yellow-400">{SITE_TITLE}</span>
			</span>
			!
		</h1>
		<h2 class="mb-4 text-gray-700 dark:text-gray-200">
			A GUI to make backup easily with
			<span class="font-semibold">Restic.</span>
		</h2>
		<p class="text-gray-600 dark:text-gray-400 mb-5">
			Simplify usage of
			<a href="https://restic.net/" target="_blank">Restic</a>
			with a
			<a href="https://go.dev/" target="_blank">Go</a>
			executable to run anywhere you want, the project is open source.
			<br />
			The website is made with
			<a href="https://kit.svelte.dev/" target="_blank">Svelte-Kit</a>
			based on
			<a href="https://swyxkit.netlify.app/" target="_blank">Swix template</a>
			.
			<br />
			Enjoy to use this product where you want.
		</p>
		<h2 class="text-gray-700 dark:text-gray-200">See how restic works ?</h2>
		<script src="https://asciinema.org/a/23554.js" id="asciicast-23554" async data-autoplay="true" data-size="big"></script>
	</div>
</div>
