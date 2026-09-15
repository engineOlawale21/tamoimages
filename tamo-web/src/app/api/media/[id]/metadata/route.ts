import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { z } from 'zod';
import { authenticatedMediaRequest } from '@/lib/auth/media-session';
import { SESSION_COOKIE, sessionCookieOptions } from '@/lib/auth/session';

const input = z.object({ title:z.string().trim().min(1).max(160), description:z.string().trim().max(2000), keywords:z.array(z.string().trim().min(1).max(64)).min(1).max(50), location:z.string().trim().max(200), usageType:z.enum(['creative','editorial']) });
export async function PATCH(request:Request,context:{params:Promise<{id:string}>}){
  try{const credential=(await cookies()).get(SESSION_COOKIE)?.value;if(!credential)return NextResponse.json({title:'Unauthorized',status:401},{status:401});const {id}=await context.params;const body=input.parse(await request.json());const result=await authenticatedMediaRequest(credential,`media/${encodeURIComponent(id)}/metadata`,{method:'PATCH',headers:{'content-type':'application/json'},body:JSON.stringify(body)});const response=new NextResponse(await result.response.text(),{status:result.response.status,headers:{'content-type':result.response.headers.get('content-type')??'application/json'}});response.cookies.set(SESSION_COOKIE,result.refreshToken,sessionCookieOptions());return response;}catch{return NextResponse.json({title:'Invalid metadata',status:422},{status:422});}
}
