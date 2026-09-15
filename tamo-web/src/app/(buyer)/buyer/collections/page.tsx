import { CollectionsManager } from '@/components/collections-manager';
import { Footer, Header } from '@/components/shell';
export default function CollectionsPage(){return <><Header/><main className="search-page"><section className="results-heading"><div><span className="eyebrow">Buyer workspace</span><h1>Collections and favourites</h1><p>Organise approved media for your projects.</p></div></section><CollectionsManager/></main><Footer/></>}
