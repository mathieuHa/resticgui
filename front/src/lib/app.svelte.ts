import { api, ApiError } from './api/client';
import { JobStore } from './jobs.svelte';

/** The jobs store shared by every page, fed by the event stream once logged in. */
export const jobs = new JobStore({ fetchJob: api.job });

export type Toast = { id: number; kind: 'success' | 'error' | 'info'; lines: string[] };

export const toasts: Toast[] = $state([]);
let nextToast = 1;

export function notify(message: string | string[], kind: Toast['kind'] = 'success', ms = 5000) {
  const toast = { id: nextToast++, kind, lines: Array.isArray(message) ? message : [message] };
  toasts.push(toast);
  if (ms > 0) setTimeout(() => dismiss(toast.id), ms);
}

export function dismiss(id: number) {
  const i = toasts.findIndex((t) => t.id === id);
  if (i >= 0) toasts.splice(i, 1);
}

/** Shows an error from the API, or any other error, until dismissed. */
export function notifyError(err: unknown, context?: string) {
  const lines = err instanceof ApiError ? err.lines : [err instanceof Error ? err.message : String(err)];
  notify(context ? [context, ...lines] : lines, 'error', 0);
}

/** Runs fn and reports its error, for button handlers. Returns whether it succeeded. */
export async function attempt(fn: () => Promise<unknown>, context?: string): Promise<boolean> {
  try {
    await fn();
    return true;
  } catch (err) {
    notifyError(err, context);
    return false;
  }
}
