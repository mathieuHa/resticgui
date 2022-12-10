import { defineConfig } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';

const config = defineConfig({
  plugins: [sveltekit()],
  server: {
    host: 'localhost',
    port: 3000,
    proxy: {
      '/api/v1': {
        target: 'http://localhost:3001/api/v1',
        changeOrigin: true,
      },
    },

  },
  build: {
    sourcemap: false,
  },
  resolve: {
    alias: {
      $utils: 'src/utils',
    },
  },
});

export default config;
