import type { AccountRole } from './session';

export function requiredRole(pathname: string): AccountRole | undefined {
  if (pathname === '/dashboard' || pathname.startsWith('/contributor/')) return 'contributor';
  if (pathname === '/buyer' || pathname.startsWith('/buyer/')) return 'buyer';
  return undefined;
}

export function canAccess(role: AccountRole, pathname: string): boolean {
  const required = requiredRole(pathname);
  return required === undefined || required === role;
}
