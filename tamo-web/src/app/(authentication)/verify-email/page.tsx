'use client';
import { useSearchParams } from 'next/navigation';
import { Suspense, useState } from 'react';
function VerifyEmailContent(){const token=useSearchParams().get('token');const [message,setMessage]=useState('');async function verify(){const response=await fetch('/api/auth/verify-email',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({token})});const body=await response.json();setMessage(body.message??body.detail??'Unable to verify email.');}return <main className="authForm authStandalone"><h1>Verify your email</h1><p>Confirm ownership of your email address to activate the account.</p><button className="primary" disabled={!token} onClick={verify}>Verify email</button>{message&&<p role="status">{message}</p>}</main>}
export default function VerifyEmail(){return <Suspense fallback={<main className="authStandalone" aria-busy="true">Loading…</main>}><VerifyEmailContent/></Suspense>}
