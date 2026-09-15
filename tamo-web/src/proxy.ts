import { NextRequest, NextResponse } from 'next/server';
import { canAccess } from './lib/auth/authorization';
import { refreshSession } from './lib/auth/identity-session';
import { SESSION_COOKIE, sessionCookieOptions } from './lib/auth/session';

export async function proxy(request:NextRequest){
  const token=request.cookies.get(SESSION_COOKIE)?.value;
  if(!token)return loginRedirect(request);
  try{
    const session=await refreshSession(token);
    if(!canAccess(session.user.role,request.nextUrl.pathname))return NextResponse.json({type:'about:blank',title:'Forbidden',status:403,detail:'Your account cannot access this area.'},{status:403});
    const response=NextResponse.next();
    response.cookies.set(SESSION_COOKIE,session.refreshToken,sessionCookieOptions());
    response.headers.set('x-tamo-account-role',session.user.role);
    return response;
  }catch{const response=loginRedirect(request);response.cookies.set(SESSION_COOKIE,'',{...sessionCookieOptions(),maxAge:0});return response;}
}

function loginRedirect(request:NextRequest){const login=new URL('/login',request.url);login.searchParams.set('returnTo',`${request.nextUrl.pathname}${request.nextUrl.search}`);return NextResponse.redirect(login);}
export const config={matcher:['/dashboard/:path*','/buyer/:path*','/contributor/:path*']};
