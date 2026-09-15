import { proxyMedia } from '@/lib/auth/media-bff';
type Context={params:Promise<{id:string;assetId:string}>};
export async function PUT(_:Request,{params}:Context){const value=await params;return proxyMedia(`collections/${encodeURIComponent(value.id)}/items/${encodeURIComponent(value.assetId)}`,{method:'PUT'});}
export async function DELETE(_:Request,{params}:Context){const value=await params;return proxyMedia(`collections/${encodeURIComponent(value.id)}/items/${encodeURIComponent(value.assetId)}`,{method:'DELETE'});}
