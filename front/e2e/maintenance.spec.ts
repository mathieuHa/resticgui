import { expect, test, type Page } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { mkdirSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';

import { RELIABILITY_PORT } from '../playwright.config';
import { e2eRoot } from './paths.mjs';

const root = e2eRoot(RELIABILITY_PORT);
const location = `${root}/old-format`;

// The maintenance tools on a repository made by an old restic: a check that
// reads data, the format upgrade, a repair after data was lost, and a new
// password. Each destructive tool asks first.
test.describe.configure({ mode: 'serial' });

function restic(password: string, ...args: string[]) {
  return execFileSync(process.env.RESTIC_BIN ?? 'restic', args, {
    env: { ...process.env, RESTIC_REPOSITORY: location, RESTIC_PASSWORD: password, RESTIC_CACHE_DIR: `${root}/cache` },
    stdio: ['ignore', 'pipe', 'pipe'],
  }).toString();
}

let id = 0;

async function settings(page: Page) {
  await page.goto(`/ui/repositories/${id}?tab=settings`);
  return page.getByTestId('maintenance');
}

// the run just started, once it ended
async function lastRun(page: Page, label: string) {
  await expect(page.getByText(`${label} started`)).toBeVisible();
  await page.getByRole('tab', { name: 'Runs' }).click();
  const run = page.getByTestId('run-history').getByRole('listitem').filter({ hasText: label }).first();
  await expect(run.locator('[data-status="succeeded"], [data-status="failed"]')).toBeVisible({ timeout: 30_000 });
  return run;
}

test.beforeAll(async ({ request }) => {
  const src = path.join(root, 'data/letters');
  mkdirSync(src, { recursive: true });
  writeFileSync(path.join(src, 'to-grandma.txt'), 'dear grandma');
  restic('first password', 'init', '--repository-version', '1');
  restic('first password', 'backup', '--host', 'laptop', src);
  const res = await request.post('/api/v2/repositories', {
    data: { name: 'Letters', backend: 'local', location, password: 'first password', host: 'laptop' },
  });
  expect(res.ok()).toBe(true);
  id = (await res.json()).id;
});

test('check part of the data now and on a schedule', async ({ page }) => {
  const tools = await settings(page);
  await tools.getByLabel('Data the check reads').selectOption('25%');
  await tools.getByRole('button', { name: 'Check now' }).click();
  const run = await lastRun(page, 'Check');
  await expect(run).toContainText('No errors found, 25% of the data read');

  await page.getByRole('tab', { name: 'Settings' }).click();
  const card = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Integrity check' }) });
  await card.getByRole('checkbox').check();
  await card.getByLabel('Data the scheduled check reads').selectOption('10%');
  await card.getByRole('button', { name: 'Save check' }).click();
  await expect(page.getByText('Integrity check saved')).toBeVisible();
  await page.reload();
  await expect(page.getByLabel('Data the scheduled check reads')).toHaveValue('10%');
});

test('upgrade the format after confirming', async ({ page }) => {
  const tools = await settings(page);
  await tools.getByRole('button', { name: 'Upgrade the format' }).click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toContainText('there is no way back');
  await dialog.getByRole('button', { name: 'Cancel' }).click();
  expect(JSON.parse(restic('first password', 'cat', 'config', '--json')).version).toBe(1);

  await tools.getByRole('button', { name: 'Upgrade the format' }).click();
  await dialog.getByRole('button', { name: 'Upgrade' }).click();
  await expect(await lastRun(page, 'Upgrade format')).toContainText('Upgraded to format version 2');
  expect(JSON.parse(restic('first password', 'cat', 'config', '--json')).version).toBe(2);
});

test('repair a repository that lost its data', async ({ page }) => {
  for (const dir of readdirSync(`${location}/data`)) rmSync(`${location}/data/${dir}`, { recursive: true });
  mkdirSync(`${location}/data/00`);
  let tools = await settings(page);
  await tools.getByRole('button', { name: 'Check now' }).click();
  await expect(await lastRun(page, 'Check')).toContainText(/\d+ errors?/);

  tools = await settings(page);
  await tools.getByRole('button', { name: 'Rebuild the index' }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Rebuild' }).click();
  await expect((await lastRun(page, 'Rebuild index')).locator('[data-status="succeeded"]')).toBeVisible();

  tools = await settings(page);
  await tools.getByRole('button', { name: 'Repair snapshots' }).click();
  await expect(page.getByRole('dialog')).toContainText('gone from these snapshots for good');
  await page.getByRole('dialog').getByRole('button', { name: 'Repair' }).click();
  await expect(await lastRun(page, 'Repair snapshots')).toContainText('Repaired 1 snapshot');

  tools = await settings(page);
  await tools.getByRole('button', { name: 'Check now' }).click();
  await expect(await lastRun(page, 'Check')).toContainText('No errors found');
});

test('remove stale locks', async ({ page }) => {
  const tools = await settings(page);
  await tools.getByRole('button', { name: 'Remove stale locks' }).click();
  await expect((await lastRun(page, 'Unlock')).locator('[data-status="succeeded"]')).toBeVisible();
});

test('change the password', async ({ page }) => {
  const tools = await settings(page);
  const change = tools.getByRole('button', { name: 'Change password' });
  await tools.getByLabel('New password').fill('second password');
  await tools.getByLabel('Repeat it').fill('second passwrd');
  await expect(tools.getByText('The two passwords differ.')).toBeVisible();
  await expect(change).toBeDisabled();
  await tools.getByLabel('Repeat it').fill('second password');
  await change.click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toContainText('stops working');
  await dialog.getByRole('button', { name: 'Change password' }).click();
  await expect((await lastRun(page, 'Change password')).locator('[data-status="succeeded"]')).toBeVisible();

  expect(() => restic('first password', 'snapshots')).toThrow();
  expect(JSON.parse(restic('second password', 'key', 'list', '--json'))).toHaveLength(1);
  // ResticGUI goes on opening it
  await page.getByRole('tab', { name: 'Snapshots' }).click();
  await expect(page.getByText('No snapshot yet')).toBeVisible();
  await expect(page.locator('.alert-warning')).toHaveCount(0);
});
