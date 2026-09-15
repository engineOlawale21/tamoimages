import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { authenticatedMediaRequest } from './media-session';
import { SESSION_COOKIE, sessionCookieOptions } from './session';

export async function proxyMedia(path:string,init:RequestInit={}){
  const credential=(await cookies()).get(SESSION_COOKIE)?.value;
  if(!credential)return NextResponse.json({title:'Unauthorized',status:401},{status:401});
  try{
    const result=await authenticatedMediaRequest(credential,path,init);
    const response=new NextResponse(result.response.status===204?null:await result.response.text(),{status:result.response.status,headers:{'content-type':result.response.headers.get('content-type')??'application/json'}});
    response.cookies.set(SESSION_COOKIE,result.refreshToken,sessionCookieOptions());return response;
  }catch{return NextResponse.json({title:'Session expired',status:401},{status:401});}
}
