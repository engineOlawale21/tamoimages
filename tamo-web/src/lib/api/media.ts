import { createApiClient } from './client';
export type MediaKind = 'image' | 'video' | 'illustration';
export type MediaVariant = { kind: 'thumbnail' | 'poster' | 'small' | 'medium' | 'large'; url: string; contentType: string; sizeBytes: number; width?: number; height?: number; expiresAt: string };
export type MediaItem = { id: string; kind: MediaKind; filename: string; contentType: string; sizeBytes: number; status: string; durationSeconds?: number; width?: number; height?: number; codecName?: string; failureCode?: string; title?: string; description?: string; keywords: string[]; location?: string; usageType?: 'creative'|'editorial'; processedAt?: string; variants: MediaVariant[]; createdAt: string; updatedAt: string };
export type MediaPage = { items: MediaItem[]; total: number };
export type CatalogVariant = Pick<MediaVariant, 'kind'|'url'|'contentType'|'width'|'height'|'expiresAt'>;
export type CatalogItem = { id:string; kind:MediaKind; title:string; description:string; keywords:string[]; location:string; usageType:'creative'|'editorial'; orientation:'portrait'|'landscape'|'square'; width:number; height:number; durationSeconds?:number; variants:CatalogVariant[]; createdAt:string };
export type CatalogPage = { items:CatalogItem[]; page:number; pageSize:number; total:number; totalPages:number };
export function createMediaApi(baseUrl: string, fetchImplementation?: typeof fetch, timeoutMilliseconds?: number) {
  const call = createApiClient({ baseUrl, fetchImplementation, timeoutMilliseconds });
  return { list: () => call<MediaPage>('media'), get: (id: string) => call<MediaItem>(`media/${encodeURIComponent(id)}`), searchCatalog: (parameters:URLSearchParams) => call<CatalogPage>(`catalog?${parameters}`), getCatalogItem: (id:string) => call<CatalogItem>(`catalog/${encodeURIComponent(id)}`), liveness: () => call<{ status: 'alive'; service: string }>('health/live') };
}
