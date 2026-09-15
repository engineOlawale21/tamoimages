import { describe, expect, it, vi } from 'vitest';
import { createApiClient } from './client';

describe('API client', () => {
  it('uses the configured base URL and privacy-safe defaults', async () => {
    const fetchImplementation = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ status: 'alive' }), {
      status: 200, headers: { 'content-type': 'application/json', 'x-correlation-id': 'response-1' },
    }));
    const call = createApiClient({ baseUrl: 'http://localhost:4000/api/v1', fetchImplementation, correlationIdFactory: () => 'request-1' });
    await expect(call<{ status: string }>('health/live')).resolves.toEqual({ status: 'alive' });
    expect(fetchImplementation).toHaveBeenCalledWith(new URL('http://localhost:4000/api/v1/health/live'), expect.objectContaining({ credentials: 'include', cache: 'no-store' }));
    const request = fetchImplementation.mock.calls[0][1];
    expect(new Headers(request?.headers).get('X-Correlation-ID')).toBe('request-1');
  });

  it('maps problem details and correlation IDs', async () => {
    const fetchImplementation = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({
      type: 'https://httpstatuses.io/429', title: 'Too Many Requests', status: 429, detail: 'Try later.', correlationId: 'request-2',
    }), { status: 429, headers: { 'content-type': 'application/problem+json', 'x-correlation-id': 'request-2' } }));
    const call = createApiClient({ baseUrl: 'http://localhost:4000/api/v1', fetchImplementation });
    await expect(call('auth/login')).rejects.toMatchObject({ status: 429, correlationId: 'request-2', message: 'Try later.' });
  });

  it('rejects API URLs containing credentials', () => {
    expect(() => createApiClient({ baseUrl: 'https://user:secret@example.com/api' })).toThrow('must not contain credentials');
  });

  it('preserves a correlation ID supplied by the caller', async () => {
    const fetchImplementation = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));
    const call = createApiClient({ baseUrl: 'http://localhost:4000/api/v1', fetchImplementation, correlationIdFactory: () => 'generated' });
    await call('health/live', { headers: { 'X-Correlation-ID': 'forwarded' } });
    const request = fetchImplementation.mock.calls[0][1];
    expect(new Headers(request?.headers).get('X-Correlation-ID')).toBe('forwarded');
  });
});
