import Image from 'next/image';
import Link from 'next/link';
import type { CatalogAsset } from '@/lib/media/catalog';
export function Gallery({ assets, className = '' }: { assets: CatalogAsset[]; className?: string }) {
  if (!assets.length) return <div className="empty-state"><h2>No images found</h2><p>Try changing your search or filters.</p></div>;
  return <div className={`gallery ${className}`.trim()}>{assets.map((asset) => <Link className="gallery-card" href={`/media/${asset.id}`} key={asset.id} aria-label={`View ${asset.alt}`}><Image unoptimized={asset.src.startsWith('http')} src={asset.src} alt={asset.alt} width={700} height={900} sizes="(max-width: 700px) 50vw, (max-width: 1100px) 33vw, 25vw" /><span className="gallery-card__meta">{asset.category}<b>View {asset.kind??'image'}</b></span></Link>)}</div>;
}
