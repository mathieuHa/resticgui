import type { Hook, HookInput } from '#lib/api/client.ts';

export type HookEvent = Hook['event'];
export type HookKind = Hook['kind'];

/** A hook being edited: a webhook from the server has a host but no URL. */
export type HookDraft = {
  id?: number;
  event: HookEvent;
  kind: HookKind;
  command: string;
  /** Empty keeps the URL of a saved webhook. */
  url: string;
  host: string;
  timeout: number;
};

export const EVENT_LABELS: Record<HookEvent, string> = {
  before: 'Before the backup',
  success: 'After a successful backup',
  failure: 'After a failed backup',
};

export const EVENTS: HookEvent[] = ['before', 'success', 'failure'];

export function draftsOf(hooks: Hook[]): HookDraft[] {
  return hooks.map((h) => ({ id: h.id, event: h.event, kind: h.kind, command: h.command, url: '', host: h.host, timeout: h.timeout_seconds }));
}

export function newHook(event: HookEvent, kind: HookKind): HookDraft {
  return { event, kind, command: '', url: '', host: '', timeout: 300 };
}

/** What is wrong with the hooks, one line each; empty when they can be saved. */
export function hookProblems(drafts: HookDraft[], commandsAllowed: boolean): string[] {
  const problems: string[] = [];
  drafts.forEach((d, i) => {
    const n = `Hook ${i + 1}`;
    if (d.kind === 'command') {
      if (!d.command.trim()) problems.push(`${n}: enter a command`);
      if (!commandsAllowed && !d.id) problems.push(`${n}: commands are turned off on this server (hooks.commands)`);
    } else if (!d.url.trim() && !(d.id && d.host)) {
      problems.push(`${n}: enter the webhook URL`);
    } else if (d.url.trim() && !/^https?:\/\/[^/\s]+/.test(d.url.trim())) {
      problems.push(`${n}: the webhook URL starts with http:// or https://`);
    }
    if (!Number.isInteger(d.timeout) || d.timeout < 1 || d.timeout > 86400) problems.push(`${n}: the time limit is 1 to 86400 seconds`);
  });
  return problems;
}

/** The hooks to send, in the order they run: before, success, failure. */
export function toHookInputs(drafts: HookDraft[]): HookInput[] {
  return EVENTS.flatMap((event) =>
    drafts
      .filter((d) => d.event === event)
      .map((d) => {
        const input: HookInput = { event: d.event, kind: d.kind, timeout_seconds: d.timeout };
        if (d.id) input.id = d.id;
        if (d.kind === 'command') input.command = d.command.trim();
        else if (d.url.trim()) input.url = d.url.trim();
        return input;
      }),
  );
}
