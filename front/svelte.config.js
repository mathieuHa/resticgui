import preprocess from 'svelte-preprocess';
import adapter from '@sveltejs/adapter-static';
import { mdsvex } from 'mdsvex';
import remarkAbbr from 'remark-abbr';
import rehypeSlug from 'rehype-slug';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';

const mdsvexConfig = {
	extensions: ['.svelte.md', '.md', '.svx'],
	layout: { _: './src/mdsvexlayout.svelte' },
	smartypants: { dashes: 'oldschool' },
	remarkPlugins: [
		remarkAbbr,
	],
	rehypePlugins: [
		rehypeSlug,
		[rehypeAutolinkHeadings, { behavior: 'wrap' }]
	]
};

/** @type {import('@sveltejs/kit').Config} */
const config = {
	extensions: ['.svelte', '.html', '.svx', ...mdsvexConfig.extensions],
	preprocess: [
		mdsvex(mdsvexConfig),
		preprocess({
			postcss: true
		}),
	],
	kit: {
		adapter: adapter({ pages: 'build' }),
		vite: () => ({
			server: {
				proxy: {
					'/api/v1': {
						target: 'http://localhost:3001',
						changeOrigin: true,
					},
				},
			},
		}),
	},
};

export default config;
