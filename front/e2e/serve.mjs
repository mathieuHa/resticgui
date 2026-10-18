// Starts the resticgui binary for the end-to-end tests, in a fresh folder
// with a small tree to back up. Usage: node e2e/serve.mjs PORT [PASSWORD]
import { execFileSync, spawn } from 'node:child_process';
import { mkdirSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';

import { e2eRoot } from './paths.mjs';

const [port = '3001', password] = process.argv.slice(2);
const bin = path.resolve(process.env.RESTICGUI_BIN ?? '../resticgui');
const restic = process.env.RESTIC_BIN ?? 'restic';
const root = e2eRoot(port);

rmSync(root, { recursive: true, force: true });
mkdirSync(path.join(root, 'data/docs'), { recursive: true });
mkdirSync(path.join(root, 'restore'));
writeFileSync(path.join(root, 'data/docs/hello.txt'), 'hello from resticgui\n');
writeFileSync(path.join(root, 'data/notes.md'), '# notes\n');
writeFileSync(path.join(root, 'data/big.bin'), Buffer.alloc(4 * 1024 * 1024, 7));

let auth = '';
if (password) {
  const hash = execFileSync(bin, ['hash-password'], { input: password + '\n' }).toString().trim();
  auth = `auth: {username: admin, password_hash: "${hash}"}\n`;
}
writeFileSync(
  path.join(root, 'config.yaml'),
  `server: {ip: 127.0.0.1, port: ${port}}
store: {path: ${root}/resticgui.sqlite, key_file: ${root}/resticgui.key, legacy_path: ""}
restic: {bin: ${restic}}
paths: {roots: [${root}]}
logs: {level: 1}
${auth}`,
);

const child = spawn(bin, [], { cwd: root, stdio: 'inherit' });
for (const signal of /** @type {const} */ (['SIGINT', 'SIGTERM'])) process.on(signal, () => child.kill(signal));
child.on('exit', (code) => process.exit(code ?? 0));
