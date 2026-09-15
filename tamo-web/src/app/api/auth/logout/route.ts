import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { logoutSession } from '@/lib/auth/identity-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';
export async function POST(){const store=await cookies();const token=store.get(SESSION_COOKIE)?.value;if(token)await logoutSession(token).catch(()=>undefined);const response=NextResponse.json({message:'Signed out.'});response.cookies.set(SESSION_COOKIE,'',{...sessionCookieOptions(),maxAge:0});return response;}
