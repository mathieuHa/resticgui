// The screenshots of the README and the documentation, taken from the real
// application with made-up data: a demo user's folders, three repositories,
// a few backups. Run it in the Playwright container, where it can write to
// /home/demo and /srv/restic: make screenshots (it writes ../images/ui-*.png).
import { execFileSync, spawn } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { mkdirSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { chromium } from '@playwright/test';

const bin = path.resolve(process.env.RESTICGUI_BIN ?? '../build/resticgui-e2e');
const restic = process.env.RESTIC_BIN ?? 'restic';
const out = path.resolve(process.env.SCREENSHOTS_DIR ?? '../images');
const home = '/home/demo';
const repos = '/srv/restic';
const state = '/srv/resticgui';
const port = 3050;
const base = `http://127.0.0.1:${port}`;
const api = `${base}/api/v2`;
const password = 'demo repository password';

/** @param {string} name @param {number | string} bytes size of random data, or the content */
function file(name, bytes) {
  mkdirSync(path.dirname(name), { recursive: true });
  writeFileSync(name, typeof bytes === 'number' ? randomBytes(bytes) : bytes);
}

/** A demo user's folders, with no personal data. */
function folders() {
  for (const dir of [home, repos, state]) rmSync(dir, { recursive: true, force: true });
  for (const [i, place] of ['beach', 'forest', 'harbour', 'mountain', 'old-town', 'sunset'].entries())
    file(`${home}/Photos/2026/summer/${place}.jpg`, 900_000 + i * 150_000);
  for (const place of ['market', 'river', 'snow']) file(`${home}/Photos/2025/winter/${place}.jpg`, 1_200_000);
  file(`${home}/Documents/taxes/2025-return.pdf`, 240_000);
  file(`${home}/Documents/house/insurance.pdf`, 180_000);
  file(`${home}/Documents/house/lease.pdf`, 320_000);
  file(`${home}/Documents/notes/ideas.md`, '# Ideas\n\n- a garden\n- a bike trip\n');
  file(`${home}/Documents/notes/books.md`, '# Books to read\n');
  file(`${home}/Projects/website/index.html`, '<!doctype html><title>Hello</title>\n');
  file(`${home}/Projects/website/style.css`, 'body { font: 16px sans-serif }\n');
  file(`${home}/Projects/website/hero.webp`, 350_000);
  file(`${home}/Projects/website/node_modules/lib/index.js`, 400_000);
  file(`${home}/Projects/garden-planner/main.go`, 'package main\n\nfunc main() {}\n');
  for (const album of ['first-album', 'live']) for (const n of [1, 2, 3]) file(`${home}/Music/${album}/track-${n}.flac`, 600_000);
  mkdirSync(`${home}/Restored`, { recursive: true });
  mkdirSync(state, { recursive: true });
  writeFileSync(
    `${state}/config.yaml`,
    `server: {ip: 127.0.0.1, port: ${port}}
store: {path: ${state}/resticgui.sqlite, key_file: ${state}/resticgui.key, legacy_path: ""}
restic: {bin: ${restic}}
paths: {roots: [${home}, ${repos}]}
hooks: {commands: true}
logs: {level: 2}
`,
  );
}

/** A repository made with restic on the command line, to connect. */
function existing() {
  const env = { ...process.env, RESTIC_REPOSITORY: `${repos}/music`, RESTIC_PASSWORD: password };
  const run = (/** @type {string[]} */ ...args) => execFileSync(restic, args, { env, stdio: 'ignore' });
  run('init');
  run('backup', '--host', 'nas', `${home}/Music`);
  file(`${home}/Music/live/track-4.flac`, 600_000);
  run('backup', '--host', 'nas', `${home}/Music`);
}

/** @param {string} method @param {string} url @param {object} [body] @returns {Promise<any>} */
async function call(method, url, body) {
  const res = await fetch(api + url, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error(`${method} ${url}: ${res.status} ${await res.text()}`);
  return res.status === 204 ? null : res.json();
}

/** @param {number} id @param {string} kind @param {number} [planId] */
async function job(id, kind, planId) {
  const { id: jobId } = await call('POST', `/repositories/${id}/jobs`, { kind, plan_id: planId });
  for (;;) {
    const j = await call('GET', `/jobs/${jobId}`);
    if (j.status === 'succeeded') return j;
    if (!['queued', 'running'].includes(j.status)) throw new Error(`${kind} ${j.status}: ${j.error}`);
    await new Promise((r) => setTimeout(r, 200));
  }
}

const retention = { keep_daily: 7, keep_weekly: 4, keep_monthly: 12, prune: true };

async function data() {
  const repository = (/** @type {string} */ name, extra = {}) =>
    call('POST', '/repositories', { name, backend: 'local', location: `${repos}/${name.toLowerCase()}`, password, host: 'laptop', ...extra });
  const plan = (/** @type {number} */ id, /** @type {object} */ p) => call('POST', `/repositories/${id}/plans`, p);

  const photos = await repository('Photos');
  await job(photos.id, 'init');
  const pictures = await plan(photos.id, {
    name: 'Default',
    paths: { include: [`${home}/Photos`], exclude: ['*.tmp'] },
    schedules: [
      { job: 'backup', cron: '30 2 * * *' },
      { job: 'forget', cron: '0 4 * * 0' },
    ],
    retention,
  });
  await job(photos.id, 'backup', pictures.id);
  file(`${home}/Photos/2026/autumn/vineyard.jpg`, 1_400_000);
  file(`${home}/Photos/2026/autumn/orchard.jpg`, 1_100_000);
  await job(photos.id, 'backup', pictures.id);
  file(`${home}/Photos/2026/summer/beach.jpg`, 950_000);
  rmSync(`${home}/Photos/2025/winter/snow.jpg`);
  await job(photos.id, 'backup', pictures.id);

  const documents = await repository('Documents');
  await job(documents.id, 'init');
  const docs = await plan(documents.id, {
    name: 'Documents',
    paths: { include: [`${home}/Documents`], exclude: [] },
    schedules: [
      { job: 'backup', cron: '0 * * * *' },
      { job: 'forget', cron: '0 4 * * 0' },
    ],
    retention: { keep_hourly: 24, ...retention },
  });
  const code = await plan(documents.id, {
    name: 'Projects',
    tags: ['code'],
    paths: { include: [`${home}/Projects`], exclude: ['node_modules'] },
    options: { compression: 'max', exclude_caches: true },
    schedules: [{ job: 'backup', cron: '@every 6h' }],
  });
  await job(documents.id, 'backup', code.id);
  file(`${home}/Projects/website/about.html`, '<!doctype html><title>About</title>\n');
  await job(documents.id, 'backup', code.id);
  for (let i = 0; i < 4; i++) {
    file(`${home}/Documents/notes/day-${i}.md`, `# Day ${i}\n`);
    file(`${home}/Documents/scans/scan-${i}.pdf`, 150_000 * (i + 1));
    await job(documents.id, 'backup', docs.id);
  }
  await job(documents.id, 'check');

  await repository('Offsite copy', { backend: 'sftp', location: 'sftp:backup@offsite.example.com:/srv/restic/laptop' });

  await call('PUT', `/repositories/${photos.id}/healthcheck`, { url: 'https://hc-ping.com/00000000-0000-0000-0000-000000000000' });
  await call('PUT', `/repositories/${photos.id}/hooks`, [
    { event: 'before', kind: 'command', command: 'docker exec photoprism photoprism backup -i /backup', timeout_seconds: 600 },
    { event: 'success', kind: 'webhook', url: 'https://hooks.example.com/backups', timeout_seconds: 30 },
    { event: 'failure', kind: 'command', command: 'logger -t resticgui "backup of $RESTICGUI_REPOSITORY failed"', timeout_seconds: 30 },
  ]);
  await call('POST', '/notifiers', {
    name: 'My phone',
    url: 'ntfy://ntfy.example.com/backups',
    on_failure: true,
    on_success: false,
    on_missed: true,
    enabled: true,
  });
  await call('POST', '/notifiers', {
    name: 'Team chat',
    url: 'discord://token@channel',
    on_failure: true,
    on_success: true,
    on_missed: true,
    enabled: true,
  });
  return { photos: photos.id, documents: documents.id };
}

/** @param {{ photos: number, documents: number }} ids */
async function shots({ photos, documents }) {
  const browser = await chromium.launch({ executablePath: process.env.CHROMIUM_PATH || undefined });
  const context = await browser.newContext({ viewport: { width: 1280, height: 800 }, deviceScaleFactor: 2, reducedMotion: 'reduce' });
  const page = await context.newPage();
  /** @param {string} name @param {import('@playwright/test').Page | import('@playwright/test').Locator} target */
  const shot = async (name, target = page, full = true) => {
    await page.waitForLoadState('networkidle');
    await page
      .locator('.skeleton')
      .first()
      .waitFor({ state: 'detached' })
      .catch(() => {});
    // a part of the page, in a window tall enough that the sticky header stays out of it
    if (target !== page) await page.setViewportSize({ width: 1280, height: 2400 });
    await target.screenshot({ path: `${out}/${name}.png`, ...(target === page ? { fullPage: full } : {}), animations: 'disabled' });
    if (target !== page) await page.setViewportSize({ width: 1280, height: 800 });
    console.log(`${name}.png`);
  };
  const theme = (/** @type {'light' | 'dark'} */ scheme) => page.emulateMedia({ colorScheme: scheme });
  const repo = (/** @type {number} */ id, tab = '') => `${base}/ui/repositories/${id}${tab ? `?tab=${tab}` : ''}`;

  await theme('light');
  await page.goto(`${base}/ui/`);
  await page.getByTestId('repository-card').nth(2).waitFor();
  await shot('ui-dashboard-light');

  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(`${base}/ui/`);
  await shot('ui-phone');
  await page.setViewportSize({ width: 1280, height: 800 });

  await theme('dark');
  await page.goto(repo(documents));
  await page.getByTestId('chart').nth(2).waitFor();
  await shot('ui-overview-dark');

  await theme('light');
  await page.goto(repo(documents, 'runs'));
  await page.getByTestId('job-plan').first().waitFor();
  await shot('ui-runs-light', page, false);

  await page.goto(repo(documents, 'settings'));
  // one plan, with its options open
  const projects = page.getByRole('region', { name: 'Plan Projects' });
  await projects.getByText('Backup options').click();
  await shot('ui-plans', projects);

  await page.goto(repo(photos, 'settings'));
  const plan = page.getByTestId('plan-settings').first();
  await plan.getByRole('button', { name: 'Preview' }).click();
  await page.getByTestId('retention-preview').waitFor();
  await page.getByTestId('retention-preview').scrollIntoViewIfNeeded();
  await shot('ui-retention', plan);

  await theme('dark');
  await page.goto(repo(photos, 'settings'));
  await shot('ui-hooks', page.locator('section').filter({ has: page.getByRole('heading', { name: 'Hooks' }) }));

  await theme('light');
  await page.goto(repo(photos, 'snapshots'));
  await page.getByTestId('snapshots').getByRole('button', { name: 'Browse' }).first().click();
  const files = page.getByTestId('snapshot-browser');
  for (const part of `${home}/Photos/2026/summer`.split('/').filter(Boolean)) await files.getByRole('button', { name: part, exact: true }).click();
  await files.getByText('beach.jpg').waitFor();
  await shot('ui-snapshot-browser', page, false);
  await page.getByRole('button', { name: 'Restore all…' }).click();
  await page
    .getByRole('dialog')
    .getByPlaceholder(/Restore into/)
    .fill(`${home}/Restored`);
  await shot('ui-restore-dialog', page, false);

  await page.goto(repo(photos, 'snapshots'));
  await page.getByTestId('snapshots').locator('tbody tr').first().getByRole('button', { name: 'Changes' }).click();
  await page.getByRole('dialog').getByTestId('diff-changes').waitFor();
  await shot('ui-compare', page, false);

  await page.goto(`${base}/ui/notifications`);
  await page.getByTestId('notifiers').getByRole('listitem').nth(1).waitFor();
  await shot('ui-notifications', page, false);

  // the wizard, as for a new repository
  await page.goto(`${base}/ui/repositories/new`);
  await page.getByText('Amazon S3 or compatible', { exact: true }).click();
  await page.getByLabel('Name', { exact: true }).fill('Cloud');
  await page.getByLabel('Location', { exact: true }).fill('s3:s3.eu-west-3.amazonaws.com/demo-backups');
  await page.getByLabel('Repository password').fill(password);
  await shot('ui-wizard-storage', page, false);

  await page.goto(`${base}/ui/repositories/new`);
  await page.getByLabel('Name', { exact: true }).fill('External disk');
  await page.getByLabel('Location', { exact: true }).fill(`${repos}/external`);
  await page.getByLabel('Repository password').fill(password);
  await page.getByRole('button', { name: 'Next' }).click();
  await page.getByPlaceholder('/or/type/a/path').fill(`${home}/Photos`);
  await page.getByRole('button', { name: 'Add', exact: true }).click();
  await page.getByRole('button', { name: 'Next' }).click();
  await page.getByText('Every day at 02:30').first().waitFor();
  await shot('ui-wizard-when');

  // connecting a repository made with restic fills in its folders
  await page.goto(`${base}/ui/repositories/new`);
  await page.getByRole('button', { name: 'Use an existing one' }).click();
  await page.getByLabel('Name', { exact: true }).fill('Music');
  await page.getByLabel('Location', { exact: true }).fill(`${repos}/music`);
  await page.getByLabel('Repository password').fill(password);
  await page.getByRole('button', { name: 'Next' }).click();
  await page.getByTestId('found').waitFor();
  await shot('ui-connect', page, false);

  await browser.close();
}

folders();
existing();
const server = spawn(bin, [], { cwd: state, stdio: 'inherit', env: { ...process.env, GIN_MODE: 'release' } });
try {
  for (let i = 0; ; i++) {
    try {
      // nosemgrep: typescript.react.security.react-insecure-request.react-insecure-request -- the demo server on 127.0.0.1
      if ((await fetch(`${api}/version`)).ok) break;
    } catch {
      if (i > 100) throw new Error('ResticGUI did not start');
    }
    await new Promise((r) => setTimeout(r, 100));
  }
  await shots(await data());
} finally {
  server.kill();
}
