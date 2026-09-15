import { proxyMedia } from '@/lib/auth/media-bff';
export function GET(){return proxyMedia('favourites');}
