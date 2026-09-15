import { NextResponse } from 'next/server';
import { z } from 'zod';
import { identityRequest } from '@/lib/auth/identity-session';
const input=z.object({token:z.string().min(32).max(512),password:z.string().min(12).max(128)});
export async function POST(request:Request){try{const response=await identityRequest('auth/reset-password',{method:'POST',body:JSON.stringify(input.parse(await request.json()))});const body=await response.json();return NextResponse.json(body,{status:response.status});}catch{return NextResponse.json({title:'Invalid request',status:400,detail:'Unable to reset password.'},{status:400});}}
