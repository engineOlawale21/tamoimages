import { NextResponse } from 'next/server';
import { z } from 'zod';
import { registerAccount } from '@/lib/auth/identity-session';
const input=z.object({email:z.string().email().max(254),firstName:z.string().trim().min(1).max(100),lastName:z.string().trim().min(1).max(100),password:z.string().min(12).max(128),role:z.enum(['buyer','contributor']),noticeVersion:z.literal('2026-09-09'),privacyNoticeAcknowledged:z.literal(true)});
export async function POST(request:Request){try{return NextResponse.json(await registerAccount(input.parse(await request.json())),{status:202});}catch{return NextResponse.json({type:'about:blank',title:'Registration failed',status:400,detail:'Unable to complete registration.'},{status:400});}}
