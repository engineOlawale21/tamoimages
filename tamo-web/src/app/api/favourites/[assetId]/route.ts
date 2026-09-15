import { proxyMedia } from '@/lib/auth/media-bff';
type Context={params:Promise<{assetId:string}>};
export async function PUT(_:Request,{params}:Context){return proxyMedia(`favourites/${encodeURIComponent((await params).assetId)}`,{method:'PUT'});}
export async function DELETE(_:Request,{params}:Context){return proxyMedia(`favourites/${encodeURIComponent((await params).assetId)}`,{method:'DELETE'});}
