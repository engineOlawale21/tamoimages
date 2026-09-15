import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { z } from 'zod';
import { authenticatedMediaRequest } from '@/lib/auth/media-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';

const input = z.object({ kind: z.enum(['image', 'video', 'illustration']), filename: z.string().trim().min(1).max(255), contentType: z.string().trim().min(1).max(255), sizeBytes: z.number().int().positive() });
export async function POST(request: Request) {
  try {
    const credential = (await cookies()).get(SESSION_COOKIE)?.value;
    if (!credential) return NextResponse.json({ title: 'Unauthorized', status: 401 }, { status: 401 });
    const result = await authenticatedMediaRequest(credential, 'upload-sessions', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify(input.parse(await request.json())) });
    const response = new NextResponse(await result.response.text(), { status: result.response.status, headers: { 'content-type': result.response.headers.get('content-type') ?? 'application/json' } });
    response.cookies.set(SESSION_COOKIE, result.refreshToken, sessionCookieOptions());
    return response;
  } catch { return NextResponse.json({ title: 'Upload unavailable', status: 503, detail: 'Unable to create an upload session.' }, { status: 503 }); }
}
