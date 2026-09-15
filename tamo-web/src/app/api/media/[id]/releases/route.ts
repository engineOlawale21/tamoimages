import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { z } from 'zod';
import { authenticatedMediaRequest } from '@/lib/auth/media-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';

async function forward(id:string,credential:string,init?:RequestInit){const result=await authenticatedMediaRequest(credential,`media/${encodeURIComponent(id)}/releases`,init);const response=new NextResponse(await result.response.text(),{status:result.response.status,headers:{'content-type':result.response.headers.get('content-type')??'application/json'}});response.cookies.set(SESSION_COOKIE,result.refreshToken,sessionCookieOptions());return response;}
export async function GET(_:Request,context:{params:Promise<{id:string}>}){try{const credential=(await cookies()).get(SESSION_COOKIE)?.value;if(!credential)return NextResponse.json({title:'Unauthorized',status:401},{status:401});return forward((await context.params).id,credential);}catch{return NextResponse.json({title:'Releases unavailable',status:503},{status:503});}}
export async function POST(request:Request,context:{params:Promise<{id:string}>}){try{const credential=(await cookies()).get(SESSION_COOKIE)?.value;if(!credential)return NextResponse.json({title:'Unauthorized',status:401},{status:401});const body=z.object({releaseType:z.enum(['model','property']),filename:z.string().trim().min(1).max(255),contentType:z.enum(['application/pdf','image/jpeg','image/png']),sizeBytes:z.number().int().positive()}).parse(await request.json());return forward((await context.params).id,credential,{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body)});}catch{return NextResponse.json({title:'Invalid release',status:422},{status:422});}}
