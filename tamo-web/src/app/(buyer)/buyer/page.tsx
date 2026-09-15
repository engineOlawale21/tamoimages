import Link from 'next/link';
import { Gallery } from '@/components/gallery';
import { Footer, Header } from '@/components/shell';
import { galleryAssets } from '@/lib/media/catalog';
import { ProfileEditor } from '@/components/profile-editor';
export default function BuyerPage(){return <><Header/><main className="search-page"><section className="results-heading"><div><span className="eyebrow">Buyer workspace</span><h1>Your recommended collection</h1><p>License authentic visuals selected for your projects.</p></div><Link className="primary small" href="/search">Explore all</Link></section><ProfileEditor role="buyer"/><Gallery assets={galleryAssets.slice(0,4)}/></main><Footer/></>}
