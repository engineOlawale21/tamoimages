'use client';
import Link from 'next/link';
import { FormEvent, useState } from 'react';
export default function ForgotPassword(){
  const [message,setMessage]=useState('');
  async function submit(event:FormEvent<HTMLFormElement>){
    event.preventDefault();const email=String(new FormData(event.currentTarget).get('email'));
    const response=await fetch('/api/auth/forgot-password',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({email})});
    const body=await response.json();
    setMessage(body.message??'If the account is eligible, instructions will be sent.');}
    return <main className="authForm authStandalone"><h1>Reset your password</h1><p>Enter your account email. The response is the same whether or not an account exists.</p>
    <form className="authForm" onSubmit={submit}>
      <input name="email" type="email" autoComplete="email" placeholder="Email" aria-label="Email address" required/>
      <button className="primary">Send instructions</button>
    </form>{message&&<p role="status">{message}</p>}
    <Link className="text-link" href="/login">Return to login</Link></main>}
