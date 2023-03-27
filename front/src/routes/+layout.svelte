<script lang="ts">
	import '../tailwind.css';
	import Nav from '$lib/Nav.svelte';
	import Toast from '$lib/Toast.svelte';

  const SITE_URL = 'https://resticgui.io';
  const REPO_URL = 'https://gitlab.com/resticgui/resticgui'
  const SITE_TITLE = 'ResticGui';
  const SITE_DESCRIPTION = 'A smart GUI for Restic, open source.';
  const DEFAULT_OG_IMAGE = '/resticLogo.png';

  console.log(`
██████╗ ███████╗███████╗████████╗██╗ ██████╗ ██████╗ ██╗   ██╗██╗
██╔══██╗██╔════╝██╔════╝╚══██╔══╝██║██╔════╝██╔════╝ ██║   ██║██║
██████╔╝█████╗  ███████╗   ██║   ██║██║     ██║  ███╗██║   ██║██║
██╔══██╗██╔══╝  ╚════██║   ██║   ██║██║     ██║   ██║██║   ██║██║
██║  ██║███████╗███████║   ██║   ██║╚██████╗╚██████╔╝╚██████╔╝██║
╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═════╝  ╚═════╝ ╚═╝

Don't hesitate to put us a star on our gitlab repository
  `);
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

<main class="min-h-[100vh] p-8 mx-auto max-w-2xl">
  <Nav />

	<slot />
</main>
