import type { Backend, Repository, RepositoryInput } from './api/client';

export const BACKEND_LABELS: Record<string, string> = {
  local: 'Local folder',
  s3: 'Amazon S3 or compatible',
  b2: 'Backblaze B2',
  rest: 'REST server',
  sftp: 'SFTP',
  rclone: 'rclone',
  azure: 'Azure Blob Storage',
  gs: 'Google Cloud Storage',
};

/** Backends in the order of BACKEND_LABELS, the common ones first. */
export function sortBackends(backends: Backend[]): Backend[] {
  const order = Object.keys(BACKEND_LABELS);
  const rank = (name: string) => (order.includes(name) ? order.indexOf(name) : order.length);
  return [...backends].sort((a, b) => rank(a.name) - rank(b.name) || a.name.localeCompare(b.name));
}

export type RepositoryDraft = {
  name: string;
  backend: string;
  location: string;
  password: string;
  env: Record<string, string>;
  secrets: Record<string, string>;
  /** on edit, whether the secrets are being replaced */
  replaceSecrets: boolean;
  cacert: string;
  extended: string;
  host: string;
};

export function newDraft(): RepositoryDraft {
  return {
    name: '',
    backend: 'local',
    location: '',
    password: '',
    env: {},
    secrets: {},
    replaceSecrets: true,
    cacert: '',
    extended: '',
    host: '',
  };
}

export function draftOf(r: Repository): RepositoryDraft {
  return {
    name: r.name,
    backend: r.backend,
    location: r.location,
    password: '',
    env: { ...r.env },
    secrets: {},
    replaceSecrets: false,
    cacert: r.cacert ?? '',
    extended: Object.entries(r.extended)
      .map(([k, v]) => `${k}=${v}`)
      .join('\n'),
    host: r.host,
  };
}

/** Parses "key=value" lines of restic -o options. */
export function parseExtended(text: string): Record<string, string> {
  const out: Record<string, string> = {};
  for (const line of text.split('\n')) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const i = trimmed.indexOf('=');
    if (i <= 0) {
      out[trimmed] = '';
    } else {
      out[trimmed.slice(0, i).trim()] = trimmed.slice(i + 1).trim();
    }
  }
  return out;
}

/**
 * The request body for a draft. Only the variables of the chosen backend are
 * sent, and secrets only when they are being replaced.
 */
export function toInput(d: RepositoryDraft, backend: Backend | undefined): RepositoryInput {
  const pick = (vars: Record<string, string>, names: string[]) =>
    Object.fromEntries(names.filter((n) => vars[n]?.trim()).map((n) => [n, vars[n]!.trim()]));
  const input: RepositoryInput = {
    name: d.name.trim(),
    backend: d.backend as RepositoryInput['backend'],
    location: d.location.trim(),
    env: pick(d.env, backend?.env ?? []),
    cacert: d.cacert.trim(),
    extended: parseExtended(d.extended),
    host: d.host.trim(),
  };
  if (d.password) input.password = d.password;
  if (d.replaceSecrets) input.secrets = pick(d.secrets, backend?.secrets ?? []);
  return input;
}

/** Problems the server would refuse, found before sending. */
export function problems(d: RepositoryDraft, backend: Backend | undefined, creating: boolean): string[] {
  const out: string[] = [];
  if (!d.name.trim()) out.push('Give the repository a name');
  if (!d.location.trim()) out.push('Enter where the repository is');
  else if (backend?.prefix && !d.location.trim().startsWith(backend.prefix))
    out.push(`${BACKEND_LABELS[d.backend] ?? d.backend} locations start with ${backend.prefix}`);
  if (creating && !d.password) out.push('Choose the repository password');
  for (const name of backend?.required ?? []) {
    const isSecret = backend?.secrets.includes(name);
    if (isSecret && !d.replaceSecrets) continue;
    if (!(isSecret ? d.secrets[name] : d.env[name])?.trim()) out.push(`${name} is required`);
  }
  return out;
}
