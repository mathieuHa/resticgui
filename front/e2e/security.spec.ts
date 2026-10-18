import { expect, test, type Page } from '@playwright/test';
import { writeFileSync } from 'node:fs';
import path from 'node:path';

import { e2eRoot } from './paths.mjs';

// Runs after backup.spec.ts on the same server, with its Documents repository.
test.describe.configure({ mode: 'serial' });

/** Fails the test on any script dialog or Content-Security-Policy report. */
function watch(page: Page) {
  const problems: string[] = [];
  page.on('dialog', (d) => {
    problems.push(`dialog: ${d.message()}`);
    void d.dismiss();
  });
  page.on('console', (m) => {
    if (/Content Security Policy|Refused to/i.test(m.text()))
      problems.push(m.text());
  });
  page.on('pageerror', (e) => problems.push(`error: ${e.message}`));
  return problems;
}

const hostile = '<img src=x onerror=alert(1)>.txt';

test('a hostile file name renders as text', async ({ page }) => {
  const problems = watch(page);
  const root = e2eRoot(new URL(test.info().project.use.baseURL!).port);
  writeFileSync(path.join(root, 'data', hostile), 'not markup\n');

  await page.goto('/ui/');
  const card = page
    .getByTestId('repository-card')
    .filter({ hasText: 'Documents' });
  await card.getByRole('button', { name: 'Back up now' }).click();
  await expect(card.getByRole('button', { name: 'Back up now' })).toBeEnabled({
    timeout: 30_000,
  });

  await page.getByRole('link', { name: 'Documents' }).first().click();
  await page.getByRole('tab', { name: 'Snapshots' }).click();
  await page
    .getByTestId('snapshots')
    .locator('tbody tr')
    .first()
    .getByRole('button', { name: 'Browse' })
    .click();
  const browser = page.getByTestId('snapshot-browser');
  for (const part of `${root}/data`.split('/').filter(Boolean)) {
    await browser.getByRole('button', { name: part, exact: true }).click();
  }
  await expect(browser.getByText(hostile, { exact: true })).toBeVisible();
  await expect(browser.locator('img[src="x"]')).toHaveCount(0);
  expect(problems).toEqual([]);
});

test('every page runs under the Content-Security-Policy', async ({ page }) => {
  const problems = watch(page);
  const response = await page.goto('/ui/');
  const csp = response!.headers()['content-security-policy'];
  expect(csp).toContain("frame-ancestors 'none'");
  expect(csp).not.toMatch(/script-src[^;]*unsafe/);
  expect(response!.headers()['x-frame-options']).toBe('DENY');

  await expect(page.getByTestId('repository-card').first()).toBeVisible();
  await page.getByRole('link', { name: 'Documents' }).first().click();
  for (const tab of ['Snapshots', 'Runs', 'Settings']) {
    await page.getByRole('tab', { name: tab }).click();
  }
  await page.goto('/ui/notifications');
  await expect(page.getByRole('heading').first()).toBeVisible();
  await page.goto('/ui/repositories/new');
  await expect(page.getByLabel('Name', { exact: true })).toBeVisible();
  expect(problems).toEqual([]);
});

test('another site cannot drive the API', async ({ request }) => {
  const jobs = async () =>
    (await (await request.get('/api/v2/jobs?repository=1')).json()).length;
  const before = await jobs();
  // what a browser sends for fetch(url, {method: 'POST', mode: 'no-cors', body})
  // from a page elsewhere: no preflight, so the server must refuse it
  const res = await request.post('/api/v2/repositories/1/jobs', {
    headers: {
      'Content-Type': 'text/plain;charset=UTF-8',
      Origin: 'https://evil.example',
      'Sec-Fetch-Site': 'cross-site',
    },
    data: JSON.stringify({ kind: 'backup' }),
  });
  expect(res.status()).toBe(403);
  // and through a domain that resolves to this machine (DNS rebinding)
  const rebound = await request.get('/api/v2/repositories', {
    headers: { Host: 'rebind.evil.example' },
  });
  expect(rebound.status()).toBe(421);
  expect(await jobs()).toBe(before);
});
