import { z } from 'zod';

const schema = z.object({ NEXT_PUBLIC_IDENTITY_API: z.string().url(), NEXT_PUBLIC_MEDIA_API: z.string().url(), NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN: z.string().url() });
export type PublicEnvironment = z.infer<typeof schema>;
export function publicEnvironment(source: Record<string, string | undefined> = process.env): PublicEnvironment {
  return schema.parse({ NEXT_PUBLIC_IDENTITY_API: source.NEXT_PUBLIC_IDENTITY_API, NEXT_PUBLIC_MEDIA_API: source.NEXT_PUBLIC_MEDIA_API, NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN: source.NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN });
}
