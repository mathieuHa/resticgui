import { describe, expect, it } from 'vitest';

import { changes, counts } from './diff.ts';

describe('diff', () => {
  it('lists added, removed and changed files without the noise', () => {
    const list = changes([
      { path: '/home/', modifier: 'U' },
      { path: '/home/me/b.txt', modifier: 'M' },
      { path: '/home/me/a.txt', modifier: '-' },
      { path: '/home/me/new/', modifier: '+' },
      { path: '/home/me/new/c.txt', modifier: '+' },
      { path: '/home/me/new/d.txt', modifier: '+' },
      { path: '/home/me/link', modifier: 'T' },
      { path: '/home/me/old/', modifier: '-' },
      { path: '/home/me/old/x', modifier: '-' },
      { path: '/home/me/newer', modifier: '+' },
    ]);
    expect(list).toEqual([
      { path: '/home/me/a.txt', kind: 'removed', dir: false },
      { path: '/home/me/b.txt', kind: 'changed', dir: false },
      { path: '/home/me/link', kind: 'changed', dir: false },
      { path: '/home/me/new', kind: 'added', dir: true },
      { path: '/home/me/newer', kind: 'added', dir: false },
      { path: '/home/me/old', kind: 'removed', dir: true },
    ]);
    expect(counts(list)).toEqual({ added: 2, removed: 2, changed: 2 });
  });
});
