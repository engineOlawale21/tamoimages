'use client';
import { useCallback, useEffect, useState } from 'react';
import { ImageIcon, RefreshCw, Video } from 'lucide-react';
import { MediaMetadataForm } from './media-metadata-form';
import { ReleaseUploader } from './release-uploader';
import type { MediaItem, MediaPage } from '@/lib/api/media';

function preview(asset:MediaItem){return asset.variants.find(variant=>variant.kind==='thumbnail')??asset.variants.find(variant=>variant.kind==='poster')??asset.variants[0];}

export function MediaLibrary({category}:{category:string}){
  const [page,setPage]=useState<MediaPage>();const [message,setMessage]=useState('Loading your media…');const [retrying,setRetrying]=useState<string>();
  const load=useCallback(async()=>{setMessage('Loading your media…');try{const response=await fetch('/api/media',{cache:'no-store'});if(!response.ok)throw new Error();setPage(await response.json() as MediaPage);setMessage('');}catch{setMessage('Unable to load your media.');}},[]);
  useEffect(()=>{void load();window.addEventListener('media-ready',load);return()=>window.removeEventListener('media-ready',load);},[load]);
  async function retry(id:string){setRetrying(id);try{const response=await fetch(`/api/media/${encodeURIComponent(id)}/retry`,{method:'POST'});if(!response.ok)throw new Error();await load();}catch{setMessage('Unable to retry media processing.');}finally{setRetrying(undefined);}}
  const items=page?.items.filter(asset=>category==='Photos'?asset.kind==='image':category==='Videos'?asset.kind==='video':category==='Illustrations'?asset.kind==='illustration':true);
  return <section className="media-library" aria-labelledby="library-heading"><div className="library-heading"><h2 id="library-heading">Recent uploads</h2><button type="button" onClick={()=>void load()} aria-label="Refresh media library"><RefreshCw size={18}/></button></div>{message&&<p role="status">{message}</p>}{items?.length===0&&<p>Your uploaded work will appear here.</p>}{items&&items.length>0&&<div className="media-grid">{items.map(asset=>{const variant=preview(asset);return <article key={asset.id} className="media-card">{variant?variant.contentType.startsWith('video/')?<video src={variant.url} muted controls preload="metadata"/>:<img src={variant.url} alt=""/>:<div className="media-placeholder">{asset.kind==='video'?<Video/>:<ImageIcon/>}</div>}<div><b title={asset.filename}>{asset.title??asset.filename}</b><span className={`media-status media-status--${asset.status}`}>{asset.status}</span></div>{asset.status==='ready'&&<><MediaMetadataForm asset={asset} onSaved={load}/><ReleaseUploader assetId={asset.id}/></>} {asset.failureCode&&<small>Processing could not complete. <button type="button" disabled={retrying===asset.id} onClick={()=>void retry(asset.id)}>{retrying===asset.id?'Retrying…':'Retry processing'}</button></small>}</article>;})}</div>}</section>;
}
