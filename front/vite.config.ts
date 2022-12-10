import { defineConfig } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';

const config = defineConfig({
	plugins: [sveltekit()],
	server: {
		host: 'localhost',
		port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:3001',
        changeOrigin: true,
        secure: false,      
        ws: true,
      }
  }
	},
	build: {
		sourcemap: false,
	},
  resolve:{
    alias:{
      '$utils' : 'src/utils'
    },
  },
});

export default config;
