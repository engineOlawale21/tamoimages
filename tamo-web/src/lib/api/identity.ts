import { createApiClient } from './client';
export type AccountRole = 'buyer' | 'contributor';
export type Account = { id: string; email: string; firstName: string; lastName: string; role: AccountRole };
export type AuthResult = { accessToken: string; user: Account };
export type RegisterInput = { email: string; firstName: string; lastName: string; password: string; role: AccountRole; noticeVersion:string; privacyNoticeAcknowledged:boolean };
export type LoginInput = { email: string; password: string };
export type MessageResult = { message: string };
export function createIdentityApi(baseUrl: string, fetchImplementation?: typeof fetch, timeoutMilliseconds?: number) {
  const call = createApiClient({ baseUrl, fetchImplementation, timeoutMilliseconds });
  return {
    register: (input: RegisterInput) => call<MessageResult>('auth/register', { method: 'POST', body: JSON.stringify(input) }),
    login: (input: LoginInput) => call<AuthResult>('auth/login', { method: 'POST', body: JSON.stringify(input) }),
    liveness: () => call<{ status: 'alive'; service: string }>('health/live'),
  };
}
