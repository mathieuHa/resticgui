import { afterEach, describe, expect, it, vi } from 'vitest';

import { api, ApiError, apiURL, onUnauthorized } from './client';

function respond(status: number, body?: unknown) {
  const fetch = vi.fn(async () =>
    body === undefined ? new Response(null, { status }) : new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }),
  );
  vi.stubGlobal('fetch', fetch);
  return fetch;
}

afterEach(() => {
  vi.unstubAllGlobals();
  onUnauthorized(null);
});

describe('api client', () => {
  it('builds URLs under /api/v2 and leaves out unset query values', () => {
    expect(apiURL('/jobs', { repository: 3, limit: 10, offset: undefined })).toBe('/api/v2/jobs?repository=3&limit=10');
    expect(api.downloadURL(1, 'latest', '/home/me/a b.txt', 'none')).toBe(
      '/api/v2/repositories/1/snapshots/latest/download?path=%2Fhome%2Fme%2Fa+b.txt&archive=none',
    );
  });

  it('sends JSON and reads JSON', async () => {
    const fetch = respond(200, { id: 1 });
    await expect(api.setPaths(1, { include: ['/a'], exclude: [] })).resolves.toEqual({ id: 1 });
    expect(fetch).toHaveBeenCalledWith('/api/v2/repositories/1/paths', {
      method: 'PUT',
      credentials: 'same-origin',
      body: '{"include":["/a"],"exclude":[]}',
      headers: { 'Content-Type': 'application/json' },
    });
  });

  it('turns an error response into an ApiError with its details', async () => {
    respond(400, { error: 'invalid request', details: ['name is required', 'location is required'] });
    const err = await api.createRepository({ name: '', backend: 'local', location: '', compression: 'auto' }).catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err.status).toBe(400);
    expect(err.lines).toEqual(['invalid request', 'name is required', 'location is required']);
  });

  it('keeps the status line when the body is not JSON', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('bad gateway', { status: 502, statusText: 'Bad Gateway' })));
    await expect(api.repositories()).rejects.toThrow('502 Bad Gateway');
  });

  it('reports a lost session, except for the login itself', async () => {
    const lost = vi.fn();
    onUnauthorized(lost);
    respond(401, { error: 'login required' });
    await expect(api.repositories()).rejects.toThrow('login required');
    expect(lost).toHaveBeenCalledOnce();
    respond(401, { error: 'wrong username or password' });
    await expect(api.login('a', 'b')).rejects.toThrow();
    expect(lost).toHaveBeenCalledOnce();
  });

  it('returns nothing for 204', async () => {
    respond(204);
    await expect(api.logout()).resolves.toBeUndefined();
  });
});
