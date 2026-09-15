'use client';
import { FormEvent, useState } from 'react';
import type { MediaItem } from '@/lib/api/media';

export function MediaMetadataForm({ asset, onSaved }: { asset: MediaItem; onSaved: () => Promise<void> }) {
  const [open, setOpen] = useState(false); const [message, setMessage] = useState('');
  async function save(event: FormEvent<HTMLFormElement>) { event.preventDefault(); setMessage('Saving…'); const values = Object.fromEntries(new FormData(event.currentTarget).entries()); const response = await fetch(`/api/media/${encodeURIComponent(asset.id)}/metadata`, { method: 'PATCH', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ title: values.title, description: values.description, keywords: String(values.keywords ?? '').split(',').map(value => value.trim()).filter(Boolean), location: values.location, usageType: values.usageType }) }); if (!response.ok) { setMessage('Unable to save metadata.'); return; } setMessage('Saved.'); await onSaved(); setOpen(false); }
  return <div className="metadata-editor">
    <button type="button" onClick={() => setOpen(value => !value)}>{open ? 'Cancel' : asset.title ? 'Edit metadata' : 'Add metadata'}</button>{open && <form onSubmit={save}><label>Title<input name="title" maxLength={160} required defaultValue={asset.title} /></label><label>Description<textarea name="description" maxLength={2000} defaultValue={asset.description} /></label><label>Keywords<input name="keywords" required defaultValue={asset.keywords.join(', ')} placeholder="lagos, street, travel" /></label><label>Location<input name="location" maxLength={200} defaultValue={asset.location} /></label><label>Classification<select name="usageType" required defaultValue={asset.usageType ?? 'creative'}><option value="creative">Creative</option><option value="editorial">Editorial</option></select></label><button className="primary small" type="submit">Save metadata</button><span role="status">{message}</span></form>}</div>;
}
