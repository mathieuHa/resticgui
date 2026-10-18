import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit({
      // a single page app: the Go server answers every /ui route with index.html
      adapter: adapter({ pages: 'build', fallback: 'index.html', strict: true }),
      paths: { base: '/ui' },
    }),
    svelteTesting(),
  ],
  server: {
    port: 3000,
    // run the Go server (make run) for the API
    proxy: { '/api': { target: 'http://localhost:3001', changeOrigin: true } },
  },
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'jsdom',
    setupFiles: ['src/test-setup.ts'],
  },
});
