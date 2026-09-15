import { z } from 'zod';

const schema = z.object({ NODE_ENV: z.enum(['development', 'test', 'production']).default('development'), API_TIMEOUT_MS: z.coerce.number().int().min(100).max(30000).default(5000), APP_ORIGIN: z.string().url().default('http://localhost:3000') });
export type ServerEnvironment = z.infer<typeof schema>;
export function serverEnvironment(source: Record<string, string | undefined> = process.env): ServerEnvironment {
  return schema.parse({ NODE_ENV: source.NODE_ENV, API_TIMEOUT_MS: source.API_TIMEOUT_MS, APP_ORIGIN: source.APP_ORIGIN });
}
