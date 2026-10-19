import { describe, expect, it } from 'vitest';

import type { Backend, Repository } from './api/client';
import { draftOf, newDraft, parseExtended, problems, toInput } from './repository';

const s3: Backend = {
  name: 's3',
  prefix: 's3:',
  example: 's3:https://s3.amazonaws.com/bucket',
  env: ['AWS_DEFAULT_REGION', 'AWS_ACCESS_KEY_ID'],
  secrets: ['AWS_ACCESS_KEY_ID', 'AWS_SECRET_ACCESS_KEY'],
  required: ['AWS_ACCESS_KEY_ID', 'AWS_SECRET_ACCESS_KEY'],
};

describe('repository drafts', () => {
  it('sends only the chosen backend variables', () => {
    const d = {
      ...newDraft(),
      name: ' photos ',
      backend: 's3',
      location: 's3:host/bucket',
      password: 'pw',
      env: { AWS_DEFAULT_REGION: 'eu-west-3', B2_ACCOUNT_ID: 'stale' },
      secrets: { AWS_ACCESS_KEY_ID: 'id', AWS_SECRET_ACCESS_KEY: 'key', B2_ACCOUNT_KEY: 'stale' },
      extended: 's3.storage-class=STANDARD_IA\n# comment\n',
    };
    expect(toInput(d, s3)).toEqual({
      name: 'photos',
      backend: 's3',
      location: 's3:host/bucket',
      password: 'pw',
      env: { AWS_DEFAULT_REGION: 'eu-west-3' },
      secrets: { AWS_ACCESS_KEY_ID: 'id', AWS_SECRET_ACCESS_KEY: 'key' },
      cacert: '',
      extended: { 's3.storage-class': 'STANDARD_IA' },
      host: '',
    });
  });

  it('keeps the stored password and secrets when editing', () => {
    const repo = {
      id: 1,
      name: 'photos',
      backend: 's3',
      location: 's3:host/bucket',
      env: { AWS_DEFAULT_REGION: 'eu-west-3' },
      secrets: ['AWS_ACCESS_KEY_ID', 'AWS_SECRET_ACCESS_KEY'],
      has_password: true,
      extended: { 's3.connections': '8' },
      host: '',
    } as unknown as Repository;
    const d = draftOf(repo);
    expect(d.extended).toBe('s3.connections=8');
    const input = toInput(d, s3);
    expect(input).not.toHaveProperty('password');
    expect(input).not.toHaveProperty('secrets');
    expect(problems(d, s3, false)).toEqual([]);
  });

  it('lists what is missing', () => {
    const d = { ...newDraft(), backend: 's3', location: 'https://bucket' };
    expect(problems(d, s3, true)).toEqual([
      'Give the repository a name',
      'Amazon S3 or compatible locations start with s3:',
      'Choose the repository password',
      'AWS_ACCESS_KEY_ID is required',
      'AWS_SECRET_ACCESS_KEY is required',
    ]);
  });

  it('parses extended options', () => {
    expect(parseExtended('a.b = 1\n\nflag.only\nx.y=a=b')).toEqual({ 'a.b': '1', 'flag.only': '', 'x.y': 'a=b' });
  });
});
