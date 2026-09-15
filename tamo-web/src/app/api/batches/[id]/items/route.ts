import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { z } from 'zod';
import { authenticatedMediaRequest } from '@/lib/auth/media-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';

export async function POST(request: Request, context: { params: Promise<{ id: string }> }) {
  try {
    const credential = (await cookies()).get(SESSION_COOKIE)?.value;
    if (!credential) return NextResponse.json({ title: 'Unauthorized', status: 401 }, { status: 401 });
    const { id } = await context.params;
    const body = z.object({ assetId: z.string().uuid() }).parse(await request.json());
    const result = await authenticatedMediaRequest(credential, `batches/${encodeURIComponent(id)}/items`, { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify(body) });
    const response = new NextResponse(await result.response.text(), { status: result.response.status, headers: { 'content-type': result.response.headers.get('content-type') ?? 'application/json' } });
    response.cookies.set(SESSION_COOKIE, result.refreshToken, sessionCookieOptions()); return response;
  } catch { return NextResponse.json({ title: 'Unable to add media', status: 422 }, { status: 422 }); }
}
