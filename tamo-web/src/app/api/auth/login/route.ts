import { NextResponse } from 'next/server';
import { z } from 'zod';
import { loginSession } from '@/lib/auth/identity-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';
const input=z.object({email:z.string().email().max(254),password:z.string().min(1).max(128)});
export async function POST(request:Request){try{const session=await loginSession(input.parse(await request.json()));const response=NextResponse.json({user:session.user,accessToken:session.accessToken});response.cookies.set(SESSION_COOKIE,session.refreshToken,sessionCookieOptions());return response;}catch{return NextResponse.json({type:'about:blank',title:'Authentication failed',status:401,detail:'Unable to sign in.'},{status:401});}}
