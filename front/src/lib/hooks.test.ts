import { describe, expect, it } from 'vitest';

import { draftsOf, hookProblems, newHook, toHookInputs } from './hooks.ts';

describe('hooks', () => {
  it('sends hooks in the order they run, keeping saved webhook URLs', () => {
    const drafts = [
      { ...newHook('success', 'webhook'), id: 4, host: 'hooks.example.com' },
      { ...newHook('before', 'command'), command: '  pg_dump app  ' },
      { ...newHook('failure', 'webhook'), url: ' https://example.com/fail ', timeout: 30 },
    ];
    expect(toHookInputs(drafts)).toEqual([
      { event: 'before', kind: 'command', command: 'pg_dump app', timeout_seconds: 300 },
      { id: 4, event: 'success', kind: 'webhook', timeout_seconds: 300 },
      { event: 'failure', kind: 'webhook', url: 'https://example.com/fail', timeout_seconds: 30 },
    ]);
  });

  it('says what is missing', () => {
    expect(
      hookProblems(
        [
          newHook('before', 'command'),
          newHook('success', 'webhook'),
          { ...newHook('success', 'webhook'), url: 'ftp://x' },
          { ...newHook('failure', 'command'), command: 'true', timeout: 0 },
        ],
        true,
      ),
    ).toEqual([
      'Hook 1: enter a command',
      'Hook 2: enter the webhook URL',
      'Hook 3: the webhook URL starts with http:// or https://',
      'Hook 4: the time limit is 1 to 86400 seconds',
    ]);
  });

  it('refuses new commands when the server turns them off, but keeps saved ones', () => {
    const saved = draftsOf([{ id: 1, event: 'before', kind: 'command', command: 'true', host: '', timeout_seconds: 60 }]);
    expect(hookProblems(saved, false)).toEqual([]);
    expect(hookProblems([{ ...newHook('before', 'command'), command: 'true' }], false)).toEqual([
      'Hook 1: commands are turned off on this server (hooks.commands)',
    ]);
  });
});
