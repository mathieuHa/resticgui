import { defineConfig, devices } from '@playwright/test';

// Runs against the built binary with the frontend embedded (make e2e), and
// real restic. RESTICGUI_BIN and RESTIC_BIN point at them.
export const OPEN_PORT = 3097;
export const AUTH_PORT = 3096;
export const PASSWORD = 'correct horse battery';

export default defineConfig({
  testDir: 'e2e',
  timeout: 60_000,
  expect: { timeout: 15_000 },
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    ...devices['Desktop Chrome'],
    launchOptions: process.env.CHROMIUM_PATH ? { executablePath: process.env.CHROMIUM_PATH } : {},
  },
  projects: [
    { name: 'open', testMatch: 'backup.spec.ts', use: { baseURL: `http://127.0.0.1:${OPEN_PORT}` } },
    { name: 'auth', testMatch: 'login.spec.ts', use: { baseURL: `http://127.0.0.1:${AUTH_PORT}` } },
  ],
  webServer: [
    { command: `node e2e/serve.mjs ${OPEN_PORT}`, url: `http://127.0.0.1:${OPEN_PORT}/api/v2/auth/session`, reuseExistingServer: false },
    { command: `node e2e/serve.mjs ${AUTH_PORT} "${PASSWORD}"`, url: `http://127.0.0.1:${AUTH_PORT}/api/v2/auth/session`, reuseExistingServer: false },
  ],
});
