import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    // a single page app: the Go server answers every /ui route with index.html
    adapter: adapter({ pages: 'build', fallback: 'index.html', strict: true }),
    paths: { base: '/ui' },
  },
};

export default config;
