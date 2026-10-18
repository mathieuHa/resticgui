import { expect, test } from '@playwright/test';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';

import { OPEN_PORT } from '../playwright.config';
import { e2eRoot } from './paths.mjs';

const root = e2eRoot(OPEN_PORT);

// One repository through its whole life: create, back up with live progress,
// browse, restore, preview retention and forget. Steps share the server.
test.describe.configure({ mode: 'serial' });

test('create a repository with the wizard', async ({ page }) => {
  await page.goto('/ui/');
  await expect(page.getByRole('heading', { name: 'No repository yet' })).toBeVisible();
  await page.getByRole('link', { name: 'Add a repository' }).click();

  // storage: validation first
  await page.getByRole('button', { name: 'Next' }).click();
  await expect(page.getByTestId('problems')).toContainText('Give the repository a name');
  await page.getByLabel('Name', { exact: true }).fill('Documents');
  await page.getByLabel('Location', { exact: true }).fill(`${root}/repo`);
  await page.getByLabel('Repository password').fill('e2e repository password');
  await page.getByRole('button', { name: 'Next' }).click();

  // what to back up: pick a folder with the browser, type another path
  const picker = page.getByTestId('path-picker');
  // the allowed roots are listed by their full path
  await picker.getByRole('button', { name: `📁 ${root}`, exact: true }).click();
  await picker.getByRole('button', { name: '📁 data', exact: true }).click();
  await page.getByRole('button', { name: 'Add this folder' }).click();
  await expect(page.getByTestId('include-paths')).toContainText(`${root}/data`);
  await page.getByRole('button', { name: 'Next' }).click();

  // when: keep the defaults
  await expect(page.getByText('Every day at 02:30')).toBeVisible();
  await page.getByRole('button', { name: 'Next' }).click();

  await expect(page.getByText('Created now with restic init')).toBeVisible();
  await page.getByRole('button', { name: 'Create repository' }).click();

  await expect(page.getByRole('heading', { name: 'Documents' })).toBeVisible();
  await page.getByRole('tab', { name: 'Runs' }).click();
  const init = page.getByTestId('run-history').getByRole('row').filter({ hasText: 'Create repository' });
  await expect(init.locator('[data-status="succeeded"]')).toBeVisible({ timeout: 30_000 });
  expect(existsSync(path.join(root, 'repo/config'))).toBe(true);
});

test('back up now and follow the progress', async ({ page }) => {
  await page.goto('/ui/');
  const card = page.getByTestId('repository-card').filter({ hasText: 'Documents' });
  await expect(card).toContainText('No backup yet');
  await card.getByRole('button', { name: 'Back up now' }).click();
  // the card shows the running job, then the result
  await expect(card.getByTestId('job-progress').or(card.getByText(/Last backup/))).toBeVisible();
  await expect(card).toContainText(/Last backup .*(now|second)/, { timeout: 30_000 });
  await expect(card.getByRole('button', { name: 'Back up now' })).toBeEnabled();
});

test('browse and download from a snapshot', async ({ page }) => {
  await page.goto('/ui/');
  await page.getByRole('link', { name: 'Documents' }).first().click();
  await page.getByRole('tab', { name: 'Snapshots' }).click();
  const rows = page.getByTestId('snapshots').locator('tbody tr');
  await expect(rows).toHaveCount(1);
  await rows.first().getByRole('button', { name: 'Browse' }).click();

  const browser = page.getByTestId('snapshot-browser');
  // the snapshot keeps full paths: walk down to the backed up folder
  for (const part of `${root}/data`.split('/').filter(Boolean)) {
    await browser.getByRole('button', { name: `📁 ${part}`, exact: true }).click();
  }
  await expect(browser.getByText('notes.md')).toBeVisible();
  await browser.getByRole('button', { name: '📁 docs', exact: true }).click();
  const file = browser.getByRole('row').filter({ hasText: 'hello.txt' });
  const download = page.waitForEvent('download');
  await file.getByRole('link', { name: 'Download' }).click();
  const saved = await (await download).path();
  expect(readFileSync(saved, 'utf8')).toBe('hello from resticgui\n');

  // find by name
  await browser.getByPlaceholder(/Find files by name/).fill('*.md');
  await browser.getByRole('button', { name: 'Find' }).click();
  await expect(browser.getByText('1 found')).toBeVisible();
  await expect(browser.getByText(`${root}/data/notes.md`)).toBeVisible();
});

test('restore a folder', async ({ page }) => {
  await page.goto('/ui/');
  await page.getByRole('link', { name: 'Documents' }).first().click();
  await page.getByRole('tab', { name: 'Snapshots' }).click();
  await page.getByTestId('snapshots').getByRole('button', { name: 'Browse' }).click();
  const browser = page.getByTestId('snapshot-browser');
  await browser.getByPlaceholder(/Find files by name/).fill('docs');
  await browser.getByRole('button', { name: 'Find' }).click();
  // restic find also matches files inside a matching folder
  await expect(browser.getByText('2 found')).toBeVisible();
  const folder = browser.getByRole('listitem').filter({ hasText: `${root}/data/docs` }).filter({ hasNotText: 'hello.txt' });
  await folder.getByRole('button', { name: 'Restore…' }).click();

  const dialog = page.getByRole('dialog');
  await dialog.getByPlaceholder(/Restore into/).fill(`${root}/restore`);
  await dialog.getByRole('button', { name: 'Restore', exact: true }).click();
  await expect(page.getByText(/Restoring into/)).toBeVisible();

  await page.getByRole('tab', { name: 'Runs' }).click();
  const run = page.getByTestId('run-history').getByRole('row').filter({ hasText: 'Restore' });
  await expect(run.locator('[data-status="succeeded"]')).toBeVisible({ timeout: 30_000 });
  await expect(run).toContainText(`into ${root}/restore`);
  expect(readFileSync(path.join(root, 'restore', root, 'data/docs/hello.txt'), 'utf8')).toBe('hello from resticgui\n');
  expect(existsSync(path.join(root, 'restore', root, 'data/notes.md'))).toBe(false);
});

test('preview retention, then forget the snapshot', async ({ page }) => {
  await page.goto('/ui/');
  await page.getByRole('link', { name: 'Documents' }).first().click();
  await page.getByRole('tab', { name: 'Settings' }).click();
  await page.getByRole('button', { name: 'Preview' }).click();
  const preview = page.getByTestId('retention-preview');
  await expect(preview).toContainText('Keeps 1');
  await expect(preview).toContainText('Removes 0');
  await expect(preview).toContainText('daily snapshot');

  await page.getByRole('tab', { name: 'Snapshots' }).click();
  await page.getByTestId('snapshots').getByRole('button', { name: 'Forget' }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Forget' }).click();
  await expect(page.getByText('No snapshot yet.')).toBeVisible();
});
