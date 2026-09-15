'use client';
import { useState } from 'react';
import { Folder, Heart, ImageIcon, Video } from 'lucide-react';
import { BatchUploader } from '@/components/batch-uploader';
import { BatchManager } from '@/components/batch-manager';
import { Footer, Header } from '@/components/shell';
import { MediaLibrary } from '@/components/media-library';
import { ProfileEditor } from '@/components/profile-editor';

const tabs = [['Photos', ImageIcon], ['Videos', Video], ['Illustrations', ImageIcon], ['Batches', Folder], ['Collections', Heart]] as const;

export default function Dashboard() {
  const [tab, setTab] = useState('Photos');
  return <><Header/><div className="cover"/><main className="dashboard"><aside><div className="avatar" aria-label="Profile initials">KA</div><h2>Contributor profile</h2><p>Photographer · Videographer</p>{tabs.map(([name, Icon]) => <button type="button" className={tab === name ? 'active' : ''} onClick={() => setTab(name)} key={name}><Icon size={19}/>{name}</button>)}</aside><section className="workspace"><ProfileEditor role="contributor"/><h1>{tab}</h1>{tab === 'Batches' ? <BatchManager/> : <>{tab !== 'Collections' && <BatchUploader key={tab} category={tab}/>}<MediaLibrary category={tab}/></>}</section></main><Footer/></>;
}
