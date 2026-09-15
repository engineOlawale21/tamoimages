import type { AuthResult, LoginInput, RegisterInput } from './identity';
import type { MediaPage } from './media';

export function mockIdentityApi(overrides: Partial<{
  register: (input: RegisterInput) => Promise<AuthResult>;
  login: (input: LoginInput) => Promise<AuthResult>;
}> = {}) {
  const unavailable = async (): Promise<never> => { throw new Error('Mock response is not configured'); };
  return { register: overrides.register ?? unavailable, login: overrides.login ?? unavailable };
}

export function mockMediaApi(page: MediaPage = { items: [], total: 0 }) {
  return { list: async () => page, liveness: async () => ({ status: 'alive' as const, service: 'tamo-media-service' }) };
}
