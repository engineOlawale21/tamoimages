import { publicEnvironment } from '../env/public';
import { refreshSession } from './identity-session';

export async function authenticatedMediaRequest(refreshToken: string, path: string, init: RequestInit = {}) {
  const session = await refreshSession(refreshToken);
  const base = publicEnvironment().NEXT_PUBLIC_MEDIA_API.replace(/\/$/, '');
  const response = await fetch(`${base}/${path}`, { ...init, headers: { authorization: `Bearer ${session.accessToken}`, ...init.headers }, cache: 'no-store' });
  return { response, refreshToken: session.refreshToken };
}
