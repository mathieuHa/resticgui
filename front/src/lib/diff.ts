import type { DiffChange } from '#lib/api/client.ts';

export type ChangeKind = 'added' | 'removed' | 'changed';

const KINDS: Record<string, ChangeKind> = { '+': 'added', '-': 'removed', M: 'changed', T: 'changed', U: 'changed' };

export type Change = { path: string; kind: ChangeKind; dir: boolean };

/**
 * The changes worth listing: files and folders added or removed, files whose
 * content or type changed. A folder whose only change is that its content
 * changed is left out, its files say it better; inside an added or removed
 * folder, only the folder is kept.
 */
export function changes(list: DiffChange[]): Change[] {
  const out: Change[] = [];
  const whole: string[] = [];
  for (const c of [...list].sort((a, b) => a.path.localeCompare(b.path))) {
    const kind = KINDS[c.modifier];
    const dir = c.path.endsWith('/');
    if (!kind || (dir && kind === 'changed')) continue;
    if (whole.some((w) => c.path.startsWith(w))) continue;
    if (dir) whole.push(c.path);
    out.push({ path: dir ? c.path.slice(0, -1) : c.path, kind, dir });
  }
  return out;
}

export function counts(list: Change[]): Record<ChangeKind, number> {
  const n = { added: 0, removed: 0, changed: 0 };
  for (const c of list) n[c.kind]++;
  return n;
}
