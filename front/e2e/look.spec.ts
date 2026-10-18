import AxeBuilder from '@axe-core/playwright';
import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

import { LOOK_PORT } from '../playwright.config';
import { e2eRoot } from './paths.mjs';

const root = e2eRoot(LOOK_PORT);

// How each screen looks, in light and dark, with an accessibility check.
// The screenshots are compared only in the Playwright container
// (make e2edocker), where fonts and rendering are the same everywhere.
test.describe.configure({ mode: 'serial' });

// what changes between runs: times, durations, sizes, ids and the charts
const dynamic = (page: Page) => [page.locator('time'), page.locator('.tabular'), page.locator('.font-mono'), page.getByTestId('chart')];

async function look(page: Page, name: string) {
  const axe = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']).analyze();
  expect(
    axe.violations.map((v) => `${v.id}: ${v.nodes.map((n) => n.target.join(' ')).join(', ')}`),
    `accessibility of ${name}`,
  ).toEqual([]);
  await expect(page).toHaveScreenshot(`${name}.png`, {
    fullPage: true,
    mask: dynamic(page),
    // grey bars, not Playwright's magenta, where times, sizes and ids change
    maskColor: '#9ca3af',
  });
}

async function both(page: Page, name: string, path: string, ready: (page: Page) => Promise<void>) {
  for (const scheme of ['light', 'dark'] as const) {
    await page.emulateMedia({ colorScheme: scheme });
    await page.goto(path);
    await ready(page);
    await look(page, `${name}-${scheme}`);
  }
}

async function run(request: APIRequestContext, id: number, kind: string) {
  const job = await (await request.post(`/api/v2/repositories/${id}/jobs`, { data: { kind } })).json();
  await expect.poll(async () => (await (await request.get(`/api/v2/jobs/${job.id}`)).json()).status, { timeout: 30_000 }).toBe('succeeded');
}

async function repository(request: APIRequestContext, name: string, paths: string[]) {
  const res = await request.post('/api/v2/repositories', {
    data: {
      name,
      backend: 'local',
      location: `${root}/${name.toLowerCase()}`,
      password: 'look test password',
    },
  });
  expect(res.status()).toBe(201);
  const { id } = await res.json();
  if (paths.length)
    await request.put(`/api/v2/repositories/${id}/paths`, {
      data: { include: paths, exclude: [] },
    });
  return id as number;
}

test('first run', async ({ page }) => {
  await both(page, 'first-run', '/ui/', (p) => expect(p.getByRole('heading', { name: 'No repository yet' })).toBeVisible());
});

let photos = 0;

test('dashboard', async ({ page, request }) => {
  photos = await repository(request, 'Photos', [`${root}/data`]);
  await run(request, photos, 'init');
  await run(request, photos, 'backup');
  await run(request, photos, 'backup');
  await request.put(`/api/v2/repositories/${photos}/schedules/backup`, {
    data: { cron: '30 2 * * *', enabled: true },
  });
  const docs = await repository(request, 'Documents', []);
  await run(request, docs, 'init');

  await both(page, 'dashboard', '/ui/', (p) => expect(p.getByTestId('headline')).toHaveText('1 repository needs your attention'));

  // phone width
  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ colorScheme: 'light' });
  await page.goto('/ui/');
  await expect(page.getByTestId('repository-card')).toHaveCount(2);
  const scroll = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(scroll, 'no sideways scrolling on a phone').toBeLessThanOrEqual(0);
  await look(page, 'dashboard-phone');
});

test('repository overview, runs and snapshots', async ({ page }) => {
  const base = `/ui/repositories/${photos}`;
  await both(page, 'overview', base, (p) => expect(p.getByTestId('chart')).toHaveCount(3));
  await both(page, 'runs', `${base}?tab=runs`, (p) => expect(p.getByTestId('run-history').getByRole('listitem')).toHaveCount(3));
  await both(page, 'snapshots', `${base}?tab=snapshots`, (p) => expect(p.getByTestId('snapshots').locator('tbody tr')).toHaveCount(2));

  await page.emulateMedia({ colorScheme: 'light' });
  await page.goto(`${base}?tab=snapshots`);
  await page.getByTestId('snapshots').getByRole('button', { name: 'Browse' }).first().click();
  await expect(page.getByTestId('snapshot-browser').getByRole('button', { name: root.split('/')[1], exact: true })).toBeVisible();
  await look(page, 'snapshot-browser');
  await page.getByRole('button', { name: 'Restore all…' }).click();
  await expect(page.getByRole('dialog')).toBeVisible();
  await look(page, 'restore-dialog');

  await page.goto(`${base}?tab=snapshots`);
  await page.getByTestId('snapshots').getByRole('button', { name: 'Changes' }).click();
  await expect(page.getByRole('dialog').getByText('Nothing changed between these snapshots.')).toBeVisible();
  await look(page, 'snapshot-diff');
});

test('settings and notifications', async ({ page, request }) => {
  const res = await request.post('/api/v2/notifiers', {
    data: { name: 'My phone', url: 'ntfy://ntfy.example.com/backups', on_failure: true, on_success: false, on_missed: true, enabled: true },
  });
  expect(res.status()).toBe(201);
  await both(page, 'notifications', '/ui/notifications', (p) => expect(p.getByTestId('notifiers').getByRole('listitem')).toHaveCount(1));
  await both(page, 'settings', `/ui/repositories/${photos}?tab=settings`, (p) => expect(p.getByRole('heading', { name: 'Hooks' })).toBeVisible());
});

test('new repository wizard', async ({ page }) => {
  await both(page, 'wizard', '/ui/repositories/new', (p) => expect(p.getByRole('button', { name: 'Next' })).toBeVisible());
});

test('the theme choice sticks', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'light' });
  await page.goto('/ui/');
  await page.getByRole('radio', { name: 'Dark theme' }).click();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'resticgui-dark');
  await page.reload();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'resticgui-dark');
  await expect(page.getByRole('radio', { name: 'Dark theme' })).toHaveAttribute('aria-checked', 'true');
  await page.getByRole('radio', { name: 'System theme' }).click();
  await expect(page.locator('html')).not.toHaveAttribute('data-theme', /.+/);
});
