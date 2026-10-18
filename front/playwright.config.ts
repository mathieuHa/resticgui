import { defineConfig, devices } from '@playwright/test';

// Runs against the built binary with the frontend embedded (make e2e), and
// real restic. RESTICGUI_BIN and RESTIC_BIN point at them.
export const OPEN_PORT = 3097;
export const AUTH_PORT = 3096;
export const LOOK_PORT = 3095;
export const CALM_PORT = 3094;
export const RELIABILITY_PORT = 3093;
export const PASSWORD = 'correct horse battery';

export default defineConfig({
  testDir: 'e2e',
  timeout: 60_000,
  expect: { timeout: 15_000, toHaveScreenshot: { maxDiffPixelRatio: 0.01 } },
  // screenshots only compare inside the Playwright container (make e2edocker)
  ignoreSnapshots: !process.env.PLAYWRIGHT_SCREENSHOTS,
  snapshotPathTemplate: 'e2e/screenshots/{arg}{ext}',
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
    {
      name: 'open',
      testMatch: ['backup.spec.ts', 'security.spec.ts'],
      use: { baseURL: `http://127.0.0.1:${OPEN_PORT}` },
    },
    {
      name: 'auth',
      testMatch: 'login.spec.ts',
      use: { baseURL: `http://127.0.0.1:${AUTH_PORT}` },
    },
    {
      name: 'look',
      testMatch: 'look.spec.ts',
      use: {
        baseURL: `http://127.0.0.1:${LOOK_PORT}`,
        reducedMotion: 'reduce',
        viewport: { width: 1280, height: 900 },
      },
    },
    {
      name: 'reliability',
      testMatch: 'reliability.spec.ts',
      use: { baseURL: `http://127.0.0.1:${RELIABILITY_PORT}` },
    },
    // the main flow again with the system asking for less motion
    {
      name: 'reduced motion',
      testMatch: 'backup.spec.ts',
      use: {
        baseURL: `http://127.0.0.1:${CALM_PORT}`,
        reducedMotion: 'reduce',
      },
    },
  ],
  webServer: [
    {
      command: `node e2e/serve.mjs ${OPEN_PORT}`,
      url: `http://127.0.0.1:${OPEN_PORT}/api/v2/auth/session`,
      reuseExistingServer: false,
    },
    {
      command: `node e2e/serve.mjs ${AUTH_PORT} "${PASSWORD}"`,
      url: `http://127.0.0.1:${AUTH_PORT}/api/v2/auth/session`,
      reuseExistingServer: false,
    },
    {
      command: `node e2e/serve.mjs ${CALM_PORT}`,
      url: `http://127.0.0.1:${CALM_PORT}/api/v2/auth/session`,
      reuseExistingServer: false,
    },
    {
      command: `node e2e/serve.mjs ${RELIABILITY_PORT}`,
      url: `http://127.0.0.1:${RELIABILITY_PORT}/api/v2/auth/session`,
      reuseExistingServer: false,
    },
    {
      command: `node e2e/serve.mjs ${LOOK_PORT}`,
      url: `http://127.0.0.1:${LOOK_PORT}/api/v2/auth/session`,
      reuseExistingServer: false,
    },
  ],
});
