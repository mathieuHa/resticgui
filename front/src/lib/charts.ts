import type { Job } from './api/client';

/** One finished backup, as the charts read it. */
export type BackupPoint = {
  id: number;
  time: number;
  status: Job['status'];
  /** bytes added to the repository, before compression */
  added?: number;
  /** bytes of the backed up files */
  processed?: number;
  /** seconds */
  took?: number;
  files?: number;
};

const num = (v: unknown) => (typeof v === 'number' && Number.isFinite(v) ? v : undefined);

/** Finished backups, oldest first, at most limit of the newest. */
export function backupPoints(jobs: Job[], limit = 30): BackupPoint[] {
  return jobs
    .filter((j) => j.kind === 'backup' && j.finished_at && j.status !== 'queued' && j.status !== 'running')
    .map((j) => {
      const s = (j.summary ?? {}) as Record<string, unknown>;
      const started = j.started_at ? Date.parse(j.started_at) : NaN;
      const finished = Date.parse(j.finished_at!);
      return {
        id: j.id,
        time: finished,
        status: j.status,
        added: num(s.data_added),
        processed: num(s.total_bytes_processed),
        took: num(s.total_duration) ?? (Number.isNaN(started) ? undefined : (finished - started) / 1000),
        files: num(s.total_files_processed),
      };
    })
    .sort((a, b) => a.time - b.time)
    .slice(-limit);
}

/** A round number at or above n for an axis: 1, 2, 2.5 or 5 times a power of 10, or of 1024 for bytes. */
export function niceMax(n: number, base: 10 | 1024 = 10): number {
  if (!(n > 0)) return 1;
  if (base === 1024) {
    let unit = 1;
    while (n / unit >= 1024) unit *= 1024;
    return niceMax(n / unit) * unit;
  }
  const power = 10 ** Math.floor(Math.log10(n));
  for (const step of [1, 2, 2.5, 5, 10]) {
    if (step * power >= n) return step * power;
  }
  return 10 * power;
}

export type Bar = { x: number; y: number; width: number; height: number; point: BackupPoint; value?: number };

/**
 * Bars for value(point) in a width × height box. A point with no value (a
 * failed backup) keeps its slot with no bar, so gaps show where runs failed.
 */
export function bars(points: BackupPoint[], value: (p: BackupPoint) => number | undefined, width: number, height: number, max: number): Bar[] {
  if (!points.length) return [];
  const slot = width / points.length;
  const gap = Math.min(4, slot * 0.25);
  return points.map((point, i) => {
    const v = value(point);
    const h = v === undefined || max <= 0 ? 0 : Math.max(v > 0 ? 1 : 0, (v / max) * height);
    return { x: i * slot + gap / 2, y: height - h, width: slot - gap, height: h, point, value: v };
  });
}

/**
 * An SVG path through the points that have a value, broken where a point has
 * none, so a failed run is a gap rather than a drop to zero.
 */
export function linePath(points: BackupPoint[], value: (p: BackupPoint) => number | undefined, width: number, height: number, max: number): string {
  if (!points.length) return '';
  const step = points.length > 1 ? width / (points.length - 1) : 0;
  let d = '';
  let pen = false;
  points.forEach((p, i) => {
    const v = value(p);
    if (v === undefined || max <= 0) {
      pen = false;
      return;
    }
    const x = points.length > 1 ? i * step : width / 2;
    const y = height - (v / max) * height;
    d += `${pen ? 'L' : 'M'}${round(x)},${round(y)}`;
    pen = true;
  });
  return d;
}

const round = (n: number) => Math.round(n * 10) / 10;
