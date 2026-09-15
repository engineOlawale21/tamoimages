import { NextResponse } from 'next/server';
import { z } from 'zod';
import { identityRequest } from '@/lib/auth/identity-session';
const input=z.object({email:z.string().email().max(254)});
export async function POST(request:Request){try{const response=await identityRequest('auth/forgot-password',{method:'POST',body:JSON.stringify(input.parse(await request.json()))});const body=await response.json();return NextResponse.json(body,{status:response.status});}catch{return NextResponse.json({message:'If the account is eligible, password-reset instructions will be sent.'},{status:202});}}
