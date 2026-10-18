import { expect, test } from '@playwright/test';
import { execFileSync } from 'node:child_process';
import { mkdirSync, writeFileSync } from 'node:fs';
import { createServer, type Server } from 'node:http';
import type { AddressInfo } from 'node:net';
import path from 'node:path';

import { RELIABILITY_PORT } from '../playwright.config';
import { e2eRoot } from './paths.mjs';

const root = e2eRoot(RELIABILITY_PORT);
const password = 'made by restic itself';

// A repository made by restic on the command line, connected without init,
// its snapshots compared; then a hook that fails a backup, and the
// notification of that failure.
test.describe.configure({ mode: 'serial' });

function restic(...args: string[]) {
  return execFileSync(process.env.RESTIC_BIN ?? 'restic', args, {
    env: { ...process.env, RESTIC_REPOSITORY: `${root}/existing`, RESTIC_PASSWORD: password, RESTIC_CACHE_DIR: `${root}/cache` },
  }).toString();
}

// what webhooks and notifiers send, as the receiver got them
type Received = { path: string; body: string };
let receiver: Server;
let received: Received[] = [];
let base = '';

test.beforeAll(async () => {
  receiver = createServer((req, res) => {
    let body = '';
    req.on('data', (c) => (body += c));
    req.on('end', () => {
      received.push({ path: req.url ?? '', body });
      // the hook endpoint fails; ntfy wants a JSON answer
      if (req.url?.startsWith('/hook')) res.writeHead(500).end();
      else res.writeHead(200, { 'Content-Type': 'application/json' }).end('{"id":"e2e"}');
    });
  });
  await new Promise<void>((done) => receiver.listen(0, '127.0.0.1', done));
  base = `127.0.0.1:${(receiver.address() as AddressInfo).port}`;
});

test.afterAll(() => {
  receiver.close();
});

test('connect a repository made with restic', async ({ page }) => {
  const src = path.join(root, 'data/photos');
  mkdirSync(src, { recursive: true });
  writeFileSync(path.join(src, 'beach.jpg'), 'sand');
  writeFileSync(path.join(src, 'old.jpg'), 'gone soon');
  restic('init');
  restic('backup', '--host', 'laptop', src);
  writeFileSync(path.join(src, 'beach.jpg'), 'more sand');
  writeFileSync(path.join(src, 'forest.jpg'), 'trees');
  execFileSync('rm', [path.join(src, 'old.jpg')]);
  restic('backup', '--host', 'laptop', src);

  await page.goto('/ui/repositories/new');
  await page.getByLabel('Name', { exact: true }).fill('Photos');
  await page.getByLabel('Location', { exact: true }).fill(`${root}/existing`);
  await page.getByLabel('Repository password').fill(password);

  // creating one where one already is would fail: the wizard says so first
  await page.getByRole('button', { name: 'Next' }).click();
  await expect(page.getByTestId('problems')).toContainText('A restic repository already exists at this location');

  await page.getByRole('button', { name: 'Use an existing one' }).click();
  await page.getByRole('button', { name: 'Next' }).click();
  await expect(page.getByTestId('found')).toContainText('Found 2 snapshots');
  await expect(page.getByTestId('include-paths')).toContainText(src);
  await page.getByRole('button', { name: 'Next' }).click();
  await page.getByRole('button', { name: 'Next' }).click();
  await expect(page.getByText('Already exists, with 2 snapshots')).toBeVisible();
  await page.getByRole('button', { name: 'Add repository' }).click();

  await expect(page.getByRole('heading', { name: 'Photos' })).toBeVisible();
  await page.getByRole('tab', { name: 'Runs' }).click();
  await expect(page.getByText('Create repository')).toHaveCount(0);

  // no snapshot is tagged by ResticGUI yet, so all of them are listed
  await page.getByRole('tab', { name: 'Snapshots' }).click();
  await expect(page.getByTestId('snapshots').locator('tbody tr')).toHaveCount(2);
  await expect(page.getByRole('checkbox', { name: 'Also snapshots made outside ResticGUI' })).toBeChecked();
});

test('compare two snapshots', async ({ page }) => {
  await page.goto('/ui/');
  await page.getByRole('link', { name: 'Photos' }).first().click();
  await page.getByRole('tab', { name: 'Snapshots' }).click();
  // newest first: the second backup is what changed since the first
  await page.getByTestId('snapshots').locator('tbody tr').first().getByRole('button', { name: 'Changes' }).click();
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByTestId('diff-stats')).toContainText('1 file');
  const changes = dialog.getByTestId('diff-changes').getByRole('listitem');
  await expect(changes).toHaveCount(3);
  await expect(changes.filter({ hasText: 'beach.jpg' })).toContainText('changed');
  await expect(changes.filter({ hasText: 'forest.jpg' })).toContainText('added');
  await expect(changes.filter({ hasText: 'old.jpg' })).toContainText('removed');
  await dialog.getByRole('button', { name: 'Close', exact: true }).click();
});

test('a failing hook fails the backup, and the failure is sent', async ({ page }) => {
  await page.goto('/ui/notifications');
  await expect(page.getByText('No notifier yet.')).toBeVisible();
  await page.getByRole('button', { name: 'Add a notifier' }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByLabel('Name').fill('My phone');
  await dialog.getByLabel('URL').fill(`ntfy://${base}/backups?scheme=http`);
  await dialog.getByRole('button', { name: 'Save' }).click();
  const notifier = page.getByTestId('notifiers').getByRole('listitem');
  await expect(notifier).toContainText('ntfy');
  await expect(notifier).toContainText('Sends failures, missed backups');
  // the URL is never shown again
  await expect(page.getByText(base)).toHaveCount(0);

  await notifier.getByRole('button', { name: 'Send a test' }).click();
  await expect(page.locator('.toast').getByText('Test sent to My phone')).toBeVisible();
  await expect.poll(() => received.filter((r) => r.path === '/backups').length).toBe(1);

  await page.goto('/ui/');
  await page.getByRole('link', { name: 'Photos' }).first().click();
  await page.getByRole('tab', { name: 'Settings' }).click();
  const before = page.getByTestId('hooks-before');
  await before.getByRole('button', { name: 'Webhook' }).click();
  await before.getByLabel('Webhook URL').fill(`http://${base}/hook/before?token=secret-token`);
  await page.getByRole('button', { name: 'Save hooks' }).click();
  await expect(page.locator('.toast').getByText('Hooks saved')).toBeVisible();
  // only the host is shown, the URL carries a token
  await expect(before).toContainText('Webhook to 127.0.0.1,');
  await expect(page.getByText('secret-token')).toHaveCount(0);

  received = [];
  await page.getByRole('tab', { name: 'Overview' }).click();
  await page.getByRole('button', { name: 'Back up now' }).first().click();
  await page.getByRole('tab', { name: 'Runs' }).click();
  const run = page.getByTestId('run-history').getByRole('listitem').filter({ hasText: 'Backup' });
  await expect(run.locator('[data-status="failed"]')).toBeVisible({ timeout: 30_000 });
  await expect(run).toContainText('before hook 1 failed');

  await expect.poll(() => received.map((r) => r.path)).toEqual(['/hook/before?token=secret-token', '/backups']);
  const hook = JSON.parse(received[0]!.body);
  expect(hook).toMatchObject({ event: 'before', repository: 'Photos' });
  expect(received[1]!.body).toContain('before hook 1 failed');
});
