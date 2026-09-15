export const SESSION_COOKIE = '__Host-tamo-session';

export function sessionCookieOptions() { return {
  httpOnly: true,
  secure: process.env.NODE_ENV === 'production',
  sameSite: 'lax' as const,
  path: '/',
  maxAge: 60 * 60 * 24 * 30,
}; }

export type AccountRole = 'buyer' | 'contributor';

export function defaultRoute(role: AccountRole): string {
  return role === 'contributor' ? '/dashboard' : '/buyer';
}
