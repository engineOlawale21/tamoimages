export type ApiProblem = { type: string; title: string; status: number; detail: string; instance?: string; correlationId?: string };
export class ApiError extends Error {
  constructor(readonly status: number, readonly problem: ApiProblem, readonly correlationId?: string) { super(problem.detail); this.name = 'ApiError'; }
}
export type ApiClientOptions = {
  baseUrl: string;
  timeoutMilliseconds?: number;
  fetchImplementation?: typeof fetch;
  correlationIdFactory?: () => string;
};

export function createApiClient(options: ApiClientOptions) {
  const baseUrl = new URL(options.baseUrl);
  if (baseUrl.username || baseUrl.password) throw new Error('API base URLs must not contain credentials');
  const request = options.fetchImplementation ?? fetch;
  const timeoutMilliseconds = options.timeoutMilliseconds ?? 5000;
  return async function call<T>(path: string, init: RequestInit = {}): Promise<T> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), timeoutMilliseconds);
    const headers = new Headers(init.headers);
    headers.set('Accept', 'application/json');
    if (!headers.has('X-Correlation-ID')) {
      headers.set('X-Correlation-ID', (options.correlationIdFactory ?? (() => crypto.randomUUID()))());
    }
    if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json');
    try {
      const response = await request(new URL(path.replace(/^\//, ''), `${baseUrl.toString().replace(/\/$/, '')}/`), { ...init, headers, signal: controller.signal, credentials: 'include', cache: 'no-store' });
      const correlationId = response.headers.get('x-correlation-id') ?? undefined;
      if (!response.ok) {
        const fallback: ApiProblem = { type: 'about:blank', title: 'Request failed', status: response.status, detail: 'The request could not be completed.', correlationId };
        const problem = await response.json().catch(() => fallback) as ApiProblem;
        throw new ApiError(response.status, problem, correlationId);
      }
      if (response.status === 204) return undefined as T;
      return await response.json() as T;
    } finally { clearTimeout(timeout); }
  };
}
