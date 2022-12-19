<script lang="ts">
	import '../tailwind.css';
	import Nav from '$lib/Nav.svelte';
	import Toast from '$lib/Toast.svelte';

  const SITE_URL = 'https://resticgui.io';
  const REPO_URL = 'https://gitlab.com/resticgui/resticgui'
  const SITE_TITLE = 'ResticGui';
  const SITE_DESCRIPTION = 'Simplify usage of Restic with an Go executable to run it everywhere you want, the repository is open source.';
  const DEFAULT_OG_IMAGE = '/resticLogo.png';
</script>

<svelte:head>
	<title>ResticGUI</title>
	<meta name="theme-color" content="#facc15" media="(prefers-color-scheme: light)" />
	<meta name="theme-color" content="#3730a3" media="(prefers-color-scheme: dark)" />
  <link rel="canonical" href={SITE_URL} />
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

<div class="flex flex-col justify-center bg-gray-50 px-4 dark:bg-gray-900 sm:px-8 mb-8">
	<Nav />
</div>

<main class="min-h-[80vh] px-8 mx-auto max-w-2xl text-black dark:text-white">
	<slot />
	<div id="modals"></div>
	<Toast />
</main>

<footer class="mx-auto mb-8 flex w-full max-w-2xl flex-col items-start justify-center">
	<hr class="border-1 mb-8 w-full border-gray-200 dark:border-gray-800" />
	<div class="grid w-full max-w-2xl grid-cols-1 gap-4 px-4 pb-16 sm:grid-cols-2 sm:px-8">
		<div class="flex flex-col space-y-4">
			<span>
        <a
          class="text-gray-500 transition hover:text-gray-300"
          target="_blank"
          rel="noopener noreferrer"
          href={REPO_URL}
        >
          GitLab
        </a>
      </span>
		</div>
	</div>
</footer>
