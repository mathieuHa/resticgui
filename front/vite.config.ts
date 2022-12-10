import { defineConfig } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';

const config = defineConfig({
	plugins: [sveltekit()],
	server: {
		host: 'localhost',
		port: 3000
	},
	build: {
		sourcemap: false,
	},
  resolve:{
    alias:{
      '$utils' : 'src/utils'
    },
  }
});

export default config;
