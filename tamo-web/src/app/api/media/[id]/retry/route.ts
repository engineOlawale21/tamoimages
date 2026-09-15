import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { authenticatedMediaRequest } from '@/lib/auth/media-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';

export async function POST(_: Request, context: { params: Promise<{ id: string }> }) {
  try {
    const credential = (await cookies()).get(SESSION_COOKIE)?.value;
    if (!credential) return NextResponse.json({ title: 'Unauthorized', status: 401 }, { status: 401 });
    const { id } = await context.params;
    const result = await authenticatedMediaRequest(credential, `media/${encodeURIComponent(id)}/retry`, { method: 'POST' });
    const response = new NextResponse(await result.response.text(), { status: result.response.status, headers: { 'content-type': result.response.headers.get('content-type') ?? 'application/json' } });
    response.cookies.set(SESSION_COOKIE, result.refreshToken, sessionCookieOptions());
    return response;
  } catch { return NextResponse.json({ title: 'Retry unavailable', status: 503 }, { status: 503 }); }
}
