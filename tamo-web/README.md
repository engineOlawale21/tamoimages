# Tamo web

Next.js deployable within the Tamo Images monorepo, matching the supplied visual direction. It does not import implementation source or configuration from sibling services.

Browser and server API construction are separated under `src/lib/api`, and every outgoing request generates or forwards `X-Correlation-ID`. Protected-route middleware and role helpers provide navigation scaffolding; backend services remain authoritative for authorization. The secure cookie-session contract is documented in `docs/SESSION_SECURITY.md`, and tokens must never be stored in browser-readable persistence.

Copy `.env.example` to `.env.local`, then run:

```powershell
npm ci
npm run dev
```

`NEXT_PUBLIC_IDENTITY_API` and `NEXT_PUBLIC_MEDIA_API` are intentionally browser-visible URLs. Server-only configuration must not use the `NEXT_PUBLIC_` prefix. The build fails when either public API URL is missing or malformed.

Identity and media adapters are in `src/lib/api`. They provide bounded requests, problem-response mapping, cookie credentials, correlation-response handling, injectable fetch implementations for tests, and no-store defaults. Authentication tokens must not be persisted in browser local storage.

## Verification

```powershell
npm run lint
npm test
npm run build
```
