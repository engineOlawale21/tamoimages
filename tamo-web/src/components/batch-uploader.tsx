'use client';
import { ChangeEvent, DragEvent, useRef, useState } from 'react';
import { CheckCircle2, LoaderCircle, Upload, XCircle } from 'lucide-react';
import type { MediaItem, MediaKind } from '@/lib/api/media';

type UploadState = 'queued' | 'uploading' | 'processing' | 'ready' | 'failed';
type Entry = { key: string; file: File; state: UploadState; progress: number; message: string; asset?: MediaItem };
const accepted = new Set(['image/jpeg', 'image/png', 'image/webp', 'video/mp4', 'video/quicktime', 'video/webm']);
const wait = (milliseconds: number) => new Promise(resolve => setTimeout(resolve, milliseconds));
function kindFor(file: File, category: string): MediaKind { if (category === 'Illustrations') return 'illustration'; return file.type.startsWith('video/') ? 'video' : 'image'; }
function allowedFor(file: File, category: string) { if (!accepted.has(file.type)) return false; if (category === 'Videos') return file.type.startsWith('video/'); if (category === 'Photos' || category === 'Illustrations') return file.type.startsWith('image/'); return true; }
async function json(response: Response) { const body = await response.json(); if (!response.ok) throw new Error(body.detail ?? body.title ?? 'Request failed'); return body; }

export function BatchUploader({ category }: { category: string }) {
  const input = useRef<HTMLInputElement>(null);
  const [entries, setEntries] = useState<Entry[]>([]);
  const [busy, setBusy] = useState(false);
  const update = (key: string, change: Partial<Entry>) => setEntries(current => current.map(entry => entry.key === key ? { ...entry, ...change } : entry));
  function add(files: File[]) { setEntries(current => [...current, ...files.map(file => allowedFor(file, category) ? { key: crypto.randomUUID(), file, state: 'queued' as const, progress: 0, message: 'Ready to upload' } : { key: crypto.randomUUID(), file, state: 'failed' as const, progress: 0, message: `This file is not supported in ${category}.` })]); }
  function select(event: ChangeEvent<HTMLInputElement>) { add(Array.from(event.target.files ?? [])); event.target.value = ''; }
  function drop(event: DragEvent<HTMLDivElement>) { event.preventDefault(); add(Array.from(event.dataTransfer.files)); }
  async function process(entry: Entry) {
    try {
      update(entry.key, { state: 'uploading', progress: 10, message: 'Creating secure upload…' });
      if (!allowedFor(entry.file, category)) throw new Error(`This file is not supported in ${category}.`);
      const session = await json(await fetch('/api/media/upload-sessions', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ kind: kindFor(entry.file, category), filename: entry.file.name, contentType: entry.file.type, sizeBytes: entry.file.size }) }));
      update(entry.key, { progress: 35, message: 'Uploading original…' });
      const uploaded = await fetch(session.upload.url, { method: 'PUT', headers: { 'content-type': entry.file.type }, body: entry.file });
      if (!uploaded.ok) throw new Error('Object upload failed');
      update(entry.key, { progress: 65, message: 'Confirming upload…' });
      await json(await fetch(`/api/media/upload-sessions/${encodeURIComponent(session.asset.id)}/complete`, { method: 'POST' }));
      update(entry.key, { state: 'processing', progress: 75, message: 'Generating variants…' });
      for (let attempt = 0; attempt < 150; attempt += 1) {
        const asset = await json(await fetch(`/api/media/${encodeURIComponent(session.asset.id)}`, { cache: 'no-store' })) as MediaItem;
        if (asset.status === 'ready') { update(entry.key, { state: 'ready', progress: 100, message: `${asset.variants.length} variants ready`, asset }); window.dispatchEvent(new Event('media-ready')); return; }
        if (asset.status === 'failed') throw new Error('Media processing failed');
        await wait(2000);
      }
      throw new Error('Processing is taking longer than expected');
    } catch (error) { update(entry.key, { state: 'failed', message: error instanceof Error ? error.message : 'Upload failed' }); }
  }
  async function start() {
    setBusy(true);
    const queued = entries.filter(entry => entry.state === 'queued' || entry.state === 'failed');
    for (const entry of queued) await process(entry);
    setBusy(false);
  }
  const accept = category === 'Videos' ? 'video/mp4,video/quicktime,video/webm' : category === 'Photos' || category === 'Illustrations' ? 'image/jpeg,image/png,image/webp' : 'image/jpeg,image/png,image/webp,video/mp4,video/quicktime,video/webm';
  return <section aria-labelledby="upload-heading"><div className="drop" onDragOver={event => event.preventDefault()} onDrop={drop}><Upload aria-hidden="true"/><p>Upload <em>{category.toLowerCase()}</em> individually or in a batch</p><h2 id="upload-heading">Share your work</h2><input ref={input} className="visually-hidden" type="file" multiple accept={accept} onChange={select}/><button className="primary" type="button" onClick={() => input.current?.click()}>Choose files</button><small>JPEG, PNG, WebP, MP4, MOV, or WebM</small></div>{entries.length > 0 && <div className="upload-queue" aria-live="polite"><div className="upload-actions"><h2>{entries.length} file{entries.length === 1 ? '' : 's'} selected</h2><button className="primary small" type="button" disabled={busy || !entries.some(entry => entry.state === 'queued' || entry.state === 'failed')} onClick={start}>{busy ? 'Uploading…' : 'Start upload'}</button></div>{entries.map(entry => <article className="upload-entry" key={entry.key}><div>{entry.state === 'ready' ? <CheckCircle2/> : entry.state === 'failed' ? <XCircle/> : <LoaderCircle className={entry.state === 'queued' ? '' : 'spin'}/>}<span><b>{entry.file.name}</b><small>{entry.message}</small></span></div><progress max="100" value={entry.progress} aria-label={`${entry.file.name}: ${entry.message}`}/></article>)}</div>}</section>;
}
