import Image from 'next/image';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { Footer, Header } from '@/components/shell';
import { SaveToCollection } from '@/components/save-to-collection';
import { serverApis } from '@/lib/api/server';

export default async function MediaDetailPage({params}:{params:Promise<{id:string}>}){
  const {id}=await params;let asset;
  try{asset=await serverApis().media.getCatalogItem(id);}catch{return notFound();}
  const preview=asset.variants.find(item=>item.kind==='large'||item.kind==='medium'||item.kind==='poster'||item.kind==='thumbnail');if(!preview)return notFound();
  return <><Header/><main className="media-detail"><div className="media-detail__preview">{asset.kind==='video'&&preview.contentType.startsWith('video/')?<video src={preview.url} controls poster={asset.variants.find(item=>item.kind==='poster')?.url}/>:<Image unoptimized src={preview.url} alt={asset.title} width={preview.width||asset.width||1200} height={preview.height||asset.height||900}/>}</div><aside><span className="eyebrow">{asset.kind} · {asset.usageType}</span><h1>{asset.title}</h1><p>{asset.description}</p><dl><dt>Orientation</dt><dd>{asset.orientation}</dd><dt>Location</dt><dd>{asset.location||'Not specified'}</dd><dt>Dimensions</dt><dd>{asset.width} × {asset.height}</dd></dl><div className="keyword-row">{asset.keywords.map(keyword=><Link key={keyword} href={`/search?q=${encodeURIComponent(keyword)}`}>{keyword}</Link>)}</div><SaveToCollection assetId={asset.id}/><p>Licensing and checkout arrive in Milestone 6.</p></aside></main><Footer/></>;
}
