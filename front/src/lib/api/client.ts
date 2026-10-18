import type { components } from './schema';

type Schemas = components['schemas'];
export type Backend = Schemas['Backend'];
export type Repository = Schemas['Repository'];
export type RepositoryInput = Schemas['RepositoryInput'];
export type Paths = Schemas['Paths'];
export type Retention = Schemas['Retention'];
export type Schedule = Repository['schedules'][number];
export type ScheduledJob = Schedule['job'];
export type Job = Schemas['Job'];
export type JobKind = Job['kind'];
export type JobStatus = Job['status'];
export type JobEvent = Schemas['Event'];
export type Progress = NonNullable<JobEvent['progress']>;
export type Entry = Schemas['Entry'];
export type Snapshot = Schemas['Snapshot'];
export type FileNode = Schemas['Node'];
export type FindMatch = Schemas['FindMatch'];
export type Stats = Schemas['Stats'];
export type Session = { auth_enabled: boolean; authenticated: boolean; username?: string };
export type RetentionPreview = {
  keep: (Snapshot & { matches?: string[] })[];
  remove: Snapshot[];
};

/** Where the Go server mounts the API; the frontend itself lives under /ui. */
export const API_BASE = '/api/v2';

/** An error response of the API. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly details: string[] = [],
  ) {
    super(message);
    this.name = 'ApiError';
  }

  /** The message and each validation problem, for showing to the user. */
  get lines(): string[] {
    return [this.message, ...this.details];
  }
}

let unauthorized: (() => void) | null = null;

/** Called when a request is refused for lack of a session, to show the login page. */
export function onUnauthorized(handler: (() => void) | null) {
  unauthorized = handler;
}

type Query = Record<string, string | number | boolean | undefined>;

export function apiURL(path: string, query: Query = {}): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) params.set(key, String(value));
  }
  const qs = params.toString();
  return API_BASE + path + (qs ? `?${qs}` : '');
}

export async function request<T>(method: string, path: string, options: { body?: unknown; query?: Query } = {}): Promise<T> {
  const init: RequestInit = { method, credentials: 'same-origin', headers: {} };
  if (options.body !== undefined) {
    init.body = JSON.stringify(options.body);
    init.headers = { 'Content-Type': 'application/json' };
  }
  const res = await fetch(apiURL(path, options.query), init);
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    let details: string[] = [];
    try {
      const body = (await res.json()) as { error?: string; details?: string[] };
      if (body.error) message = body.error;
      details = body.details ?? [];
    } catch {
      // not JSON, keep the status line
    }
    if (res.status === 401 && !path.startsWith('/auth/')) unauthorized?.();
    throw new ApiError(res.status, message, details);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

const repo = (id: number) => `/repositories/${id}`;
const snap = (id: number, snapshot: string) => `${repo(id)}/snapshots/${encodeURIComponent(snapshot)}`;

export const api = {
  session: () => request<Session>('GET', '/auth/session'),
  login: (username: string, password: string) => request<void>('POST', '/auth/login', { body: { username, password } }),
  logout: () => request<void>('POST', '/auth/logout'),

  backends: () => request<Backend[]>('GET', '/backends'),
  browse: (path?: string) => request<Entry[]>('GET', '/fs', { query: { path } }),

  repositories: () => request<Repository[]>('GET', '/repositories'),
  repository: (id: number) => request<Repository>('GET', repo(id)),
  createRepository: (input: RepositoryInput) => request<Repository>('POST', '/repositories', { body: input }),
  updateRepository: (id: number, input: RepositoryInput) => request<Repository>('PUT', repo(id), { body: input }),
  deleteRepository: (id: number) => request<void>('DELETE', repo(id)),
  setPaths: (id: number, paths: Paths) => request<Repository>('PUT', `${repo(id)}/paths`, { body: paths }),
  setRetention: (id: number, retention: Retention) =>
    request<Repository>('PUT', `${repo(id)}/retention`, { body: retention }),
  deleteRetention: (id: number) => request<Repository>('DELETE', `${repo(id)}/retention`),
  previewRetention: (id: number, retention: Retention) =>
    request<RetentionPreview>('POST', `${repo(id)}/retention/preview`, { body: retention }),
  setSchedule: (id: number, job: ScheduledJob, cron: string, enabled = true) =>
    request<Repository>('PUT', `${repo(id)}/schedules/${job}`, { body: { cron, enabled } }),
  deleteSchedule: (id: number, job: ScheduledJob) => request<Repository>('DELETE', `${repo(id)}/schedules/${job}`),
  startJob: (id: number, kind: Exclude<JobKind, 'restore'>) =>
    request<Job>('POST', `${repo(id)}/jobs`, { body: { kind } }),
  stats: (id: number, snapshot?: string) => request<Stats>('GET', `${repo(id)}/stats`, { query: { snapshot } }),

  snapshots: (id: number, all = false) =>
    request<Snapshot[]>('GET', `${repo(id)}/snapshots`, { query: { all: all || undefined } }),
  forgetSnapshot: (id: number, snapshot: string, prune = false) =>
    request<void>('DELETE', snap(id, snapshot), { query: { prune: prune || undefined } }),
  files: (id: number, snapshot: string, path = '/') =>
    request<FileNode[]>('GET', `${snap(id, snapshot)}/files`, { query: { path } }),
  find: (id: number, snapshot: string, pattern: string) =>
    request<FindMatch[]>('GET', `${snap(id, snapshot)}/find`, { query: { pattern } }),
  downloadURL: (id: number, snapshot: string, path: string, archive: 'zip' | 'tar' | 'none') =>
    apiURL(`${snap(id, snapshot)}/download`, { path, archive }),
  restore: (id: number, snapshot: string, target: string, include: string[] = []) =>
    request<Job>('POST', `${snap(id, snapshot)}/restore`, { body: { target, include } }),

  jobs: (repositoryId: number, limit = 50, offset = 0) =>
    request<Job[]>('GET', '/jobs', { query: { repository: repositoryId, limit, offset } }),
  job: (id: number) => request<Job>('GET', `/jobs/${id}`),
  cancelJob: (id: number) => request<void>('POST', `/jobs/${id}/cancel`),
  eventsURL: () => apiURL('/events'),
};
