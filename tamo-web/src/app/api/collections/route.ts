import { NextResponse } from 'next/server';
import { z } from 'zod';
import { proxyMedia } from '@/lib/auth/media-bff';
const input=z.object({name:z.string().trim().min(1).max(120)}).strict();
export function GET(){return proxyMedia('collections');}
export async function POST(request:Request){try{return proxyMedia('collections',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(input.parse(await request.json()))});}catch{return NextResponse.json({title:'Invalid collection',status:400},{status:400});}}
