import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { z } from 'zod';
import { authenticatedMediaRequest } from '@/lib/auth/media-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';

async function forward(path: string, init?: RequestInit) {
  const credential = (await cookies()).get(SESSION_COOKIE)?.value;
  if (!credential) return NextResponse.json({ title: 'Unauthorized', status: 401 }, { status: 401 });
  const result = await authenticatedMediaRequest(credential, path, init);
  const response = new NextResponse(await result.response.text(), { status: result.response.status, headers: { 'content-type': result.response.headers.get('content-type') ?? 'application/json' } });
  response.cookies.set(SESSION_COOKIE, result.refreshToken, sessionCookieOptions());
  return response;
}
export async function GET() { try { return await forward('batches'); } catch { return NextResponse.json({ title: 'Batches unavailable', status: 503 }, { status: 503 }); } }
export async function POST(request: Request) { try { const body = z.object({ name: z.string().trim().min(1).max(120) }).parse(await request.json()); return await forward('batches', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify(body) }); } catch { return NextResponse.json({ title: 'Invalid batch', status: 422 }, { status: 422 }); } }
