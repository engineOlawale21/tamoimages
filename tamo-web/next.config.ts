import type { NextConfig } from 'next';
const apiOrigins = ['NEXT_PUBLIC_IDENTITY_API', 'NEXT_PUBLIC_MEDIA_API', 'NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN'].map((key) => {
  const value = process.env[key];
  if (!value) throw new Error(`${key} is required`);
  return new URL(value).origin;
});
const scriptPolicy = process.env.NODE_ENV === 'development' ? "script-src 'self' 'unsafe-inline' 'unsafe-eval'" : "script-src 'self' 'unsafe-inline'";
const securityHeaders = [
  { key: 'X-Content-Type-Options', value: 'nosniff' }, { key: 'X-Frame-Options', value: 'DENY' },
  { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' }, { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
  { key: 'Content-Security-Policy', value: `default-src 'self'; img-src 'self' data: https://images.unsplash.com; style-src 'self' 'unsafe-inline'; ${scriptPolicy}; connect-src 'self' ${apiOrigins.join(' ')}; frame-ancestors 'none'; base-uri 'self'; form-action 'self'` },
];
const config: NextConfig = {
  output: 'standalone',
  images: { remotePatterns: [{ protocol: 'https', hostname: 'images.unsplash.com' }] },
  async headers() { return [{ source: '/(.*)', headers: securityHeaders }]; },
};
export default config;
