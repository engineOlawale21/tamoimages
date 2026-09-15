import type { Metadata } from 'next';
import './styles.css';
import './batch-uploader.css';
export const metadata: Metadata = { title: 'Tamo Images — Authentic African stories', description: 'Discover and license authentic African photography, video and illustration.' };
export default function Layout({children}:{children:React.ReactNode}) { return <html lang="en"><body>{children}</body></html> }
