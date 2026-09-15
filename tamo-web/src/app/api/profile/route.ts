import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { z } from 'zod';
import { identityRequest, refreshSession } from '@/lib/auth/identity-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';

const profile=z.object({displayName:z.string().trim().max(100).optional(),organisation:z.string().trim().max(160).optional(),profession:z.string().trim().max(100).optional(),biography:z.string().trim().max(1000).optional(),countryCode:z.string().trim().regex(/^[A-Za-z]{2}$/).transform(value=>value.toUpperCase()).optional(),city:z.string().trim().max(100).optional()}).strict();
async function authorized(method:'GET'|'PATCH'|'DELETE',body?:unknown){const jar=await cookies();const credential=jar.get(SESSION_COOKIE)?.value;if(!credential)return NextResponse.json({title:'Unauthorized',status:401},{status:401});try{const session=await refreshSession(credential);const response=await identityRequest('profiles/me',{method,headers:{authorization:`Bearer ${session.accessToken}`},body:body===undefined?undefined:JSON.stringify(body)});const payload=await response.json();const result=NextResponse.json(payload,{status:response.status});result.cookies.set(SESSION_COOKIE,method==='DELETE'?'':session.refreshToken,{...sessionCookieOptions(),...(method==='DELETE'?{maxAge:0}:{})});return result;}catch{return NextResponse.json({title:'Session expired',status:401},{status:401});}}
export function GET(){return authorized('GET');}
export async function PATCH(request:Request){try{return authorized('PATCH',profile.parse(await request.json()));}catch{return NextResponse.json({title:'Invalid profile',status:400},{status:400});}}
export function DELETE(){return authorized('DELETE');}
