import { describe, expect, it } from 'vitest';
import { publicEnvironment } from './public';
import { serverEnvironment } from './server';

describe('environment boundaries', () => {
  it('exposes only explicitly selected public values', () => {
    expect(publicEnvironment({ NEXT_PUBLIC_IDENTITY_API: 'https://identity.example/api/v1', NEXT_PUBLIC_MEDIA_API: 'https://media.example/api/v1', NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN: 'https://storage.example', DATABASE_URL: 'secret' })).toEqual({
      NEXT_PUBLIC_IDENTITY_API: 'https://identity.example/api/v1', NEXT_PUBLIC_MEDIA_API: 'https://media.example/api/v1', NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN: 'https://storage.example',
    });
  });

  it('validates server-only timeout settings', () => {
    expect(serverEnvironment({ NODE_ENV: 'test', API_TIMEOUT_MS: '2500' })).toEqual({ NODE_ENV: 'test', API_TIMEOUT_MS: 2500, APP_ORIGIN: 'http://localhost:3000' });
    expect(() => serverEnvironment({ API_TIMEOUT_MS: '60000' })).toThrow();
  });
});
