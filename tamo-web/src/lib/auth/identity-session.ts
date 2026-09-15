import { createHash } from 'node:crypto';
import { publicEnvironment } from '../env/public';
import type { Account, LoginInput, RegisterInput } from '../api/identity';
import { serverEnvironment } from '../env/server';

const backendCookie='__Host-tamo-refresh';
export type IdentitySession={accessToken:string;user:Account;refreshToken:string};
const inFlightRefreshes = new Map<string, Promise<IdentitySession>>();

export async function identityRequest(path:string,init:RequestInit={}){const base=publicEnvironment().NEXT_PUBLIC_IDENTITY_API.replace(/\/$/,'');return fetch(`${base}/${path}`,{...init,headers:{'content-type':'application/json',...init.headers},cache:'no-store'});}
export function refreshTokenFrom(response:Response){const value=response.headers.get('set-cookie');const match=value?.match(new RegExp(`${backendCookie}=([^;]+)`));return match?decodeURIComponent(match[1]):undefined;}
export async function loginSession(input:LoginInput):Promise<IdentitySession>{const response=await identityRequest('auth/login',{method:'POST',body:JSON.stringify(input)});if(!response.ok)throw new Error('Unable to sign in');const body=await response.json() as Omit<IdentitySession,'refreshToken'>;const refreshToken=refreshTokenFrom(response);if(!refreshToken)throw new Error('Identity service did not create a session');return {...body,refreshToken};}
export async function registerAccount(input:RegisterInput){const response=await identityRequest('auth/register',{method:'POST',body:JSON.stringify(input)});if(!response.ok)throw new Error('Unable to register');return response.json() as Promise<{message:string}>;}
async function performRefreshSession(refreshToken:string):Promise<IdentitySession>{const response=await identityRequest('auth/refresh',{method:'POST',headers:{cookie:`${backendCookie}=${encodeURIComponent(refreshToken)}`,origin:serverEnvironment().APP_ORIGIN}});if(!response.ok)throw new Error('Session expired');const body=await response.json() as Omit<IdentitySession,'refreshToken'>;const next=refreshTokenFrom(response);if(!next)throw new Error('Identity service did not rotate the session');return {...body,refreshToken:next};}
export function refreshSession(refreshToken:string):Promise<IdentitySession>{const key=createHash('sha256').update(refreshToken).digest('base64url');const existing=inFlightRefreshes.get(key);if(existing)return existing;const pending=performRefreshSession(refreshToken);inFlightRefreshes.set(key,pending);void pending.finally(()=>inFlightRefreshes.delete(key)).catch(()=>undefined);return pending;}
export async function logoutSession(refreshToken:string){await identityRequest('auth/logout',{method:'POST',headers:{cookie:`${backendCookie}=${encodeURIComponent(refreshToken)}`,origin:serverEnvironment().APP_ORIGIN}});}
