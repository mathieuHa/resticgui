import type { Job, JobEvent, JobStatus, Progress } from './api/client';

export const FINISHED: readonly JobStatus[] = ['succeeded', 'failed', 'cancelled', 'interrupted'];
export const isFinished = (status: JobStatus) => FINISHED.includes(status);
export const isActive = (status: JobStatus) => status === 'queued' || status === 'running';

/** The part of EventSource the store uses, so that tests can feed it events. */
export interface EventSourceLike {
  addEventListener(type: string, listener: (event: MessageEvent) => void): void;
  close(): void;
  onopen: ((event: Event) => void) | null;
  onerror: ((event: Event) => void) | null;
}

type Options = {
  fetchJob: (id: number) => Promise<Job>;
  createSource?: (url: string) => EventSourceLike;
  now?: () => Date;
};

/**
 * Jobs known to the page, kept current by the server's event stream: status
 * changes of every job and the progress of running backups and restores.
 */
export class JobStore {
  /** Jobs by id. */
  byId: Record<number, Job> = $state({});
  /** Progress of running jobs by id; dropped when the job finishes. */
  progress: Record<number, Progress> = $state({});
  /** Whether the event stream is open. */
  connected = $state(false);

  #options: Options;
  #source: EventSourceLike | null = null;
  #lost = false;
  // callbacks, not state shown on a page
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  #listeners = new Set<(job: Job) => void>();
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  #resync = new Set<() => void>();

  constructor(options: Options) {
    this.#options = options;
  }

  connect(url: string) {
    this.disconnect();
    const create = this.#options.createSource ?? ((u: string) => new EventSource(u) as EventSourceLike);
    const source = create(url);
    source.onopen = () => {
      this.connected = true;
      if (this.#lost) {
        // events were missed while disconnected: let pages reload what they show
        this.#lost = false;
        for (const fn of this.#resync) fn();
      }
    };
    source.onerror = () => {
      // EventSource reconnects by itself
      this.connected = false;
      this.#lost = true;
    };
    source.addEventListener('job', (e) => this.#onJob(JSON.parse(e.data) as JobEvent));
    source.addEventListener('progress', (e) => this.#onProgress(JSON.parse(e.data) as JobEvent));
    this.#source = source;
  }

  disconnect() {
    this.#source?.close();
    this.#source = null;
    this.connected = false;
  }

  /** Adds or replaces jobs read from the API. */
  add(jobs: Job[]) {
    for (const job of jobs) {
      this.byId[job.id] = job;
      if (isFinished(job.status)) delete this.progress[job.id];
    }
  }

  /** Jobs of a repository, newest first. */
  list(repositoryId: number): Job[] {
    return Object.values(this.byId)
      .filter((j) => j.repository_id === repositoryId)
      .sort((a, b) => b.id - a.id);
  }

  latest(repositoryId: number, kind?: Job['kind']): Job | undefined {
    return this.list(repositoryId).find((j) => !kind || j.kind === kind);
  }

  /** The running job of a repository, or else its oldest queued one. */
  active(repositoryId: number): Job | undefined {
    const jobs = this.list(repositoryId);
    return jobs.find((j) => j.status === 'running') ?? jobs.filter((j) => j.status === 'queued').at(-1);
  }

  /** Calls fn whenever a job changes status. Returns the unsubscribe function. */
  onChange(fn: (job: Job) => void): () => void {
    this.#listeners.add(fn);
    return () => this.#listeners.delete(fn);
  }

  /** Calls fn after the stream reconnects, when events may have been missed. */
  onResync(fn: () => void): () => void {
    this.#resync.add(fn);
    return () => this.#resync.delete(fn);
  }

  async #onJob(event: JobEvent) {
    // eslint-disable-next-line svelte/prefer-svelte-reactivity -- a timestamp, not state
    const now = (this.#options.now?.() ?? new Date()).toISOString();
    let job = this.byId[event.job_id];
    if (job) {
      job.status = event.status as JobStatus;
      if (event.status === 'running' && !job.started_at) job.started_at = now;
      if (event.snapshot_id) job.snapshot_id = event.snapshot_id;
      if (event.error) job.error = event.error;
      if (isFinished(job.status) && !job.finished_at) job.finished_at = now;
    }
    if (isFinished(event.status as JobStatus)) delete this.progress[event.job_id];
    if (!job || isFinished(job.status)) {
      // a new job, or a finished one whose summary is now recorded
      try {
        job = await this.#options.fetchJob(event.job_id);
        this.add([job]);
      } catch {
        if (!job) return;
      }
    }
    const current = this.byId[event.job_id];
    if (current) for (const fn of this.#listeners) fn(current);
  }

  #onProgress(event: JobEvent) {
    if (!event.progress) return;
    const job = this.byId[event.job_id];
    if (job && isFinished(job.status)) return;
    this.progress[event.job_id] = event.progress;
  }
}
