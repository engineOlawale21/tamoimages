# Milestone 2 status

Last verified: 8 September 2026.

## Completed and verified

- Authentication lifecycle schema migrations apply, roll back, and reapply on local PostgreSQL.
- Registration, email verification, generic resend/recovery responses, login, rotating refresh sessions, replay-family revocation, current/all-session logout, password reset, and account-deletion state are implemented.
- Buyer and contributor profile storage is role-specific and field-minimised.
- Authenticated NDPA/NDPR rights-request receipts cover access, correction, portability, deletion, restriction, objection, consent withdrawal, and complaint workflows.
- Privacy-safe security audit writes are connected to authentication lifecycle events.
- Next.js owns the opaque HttpOnly session cookie and provides login, registration, verification, recovery, refresh validation, and logout boundaries.
- Backend lint, 15 unit tests, 7 live PostgreSQL/Redis/HTTP integration tests, build, OpenAPI export, and contract validation pass.
- Frontend lint, 10 unit tests, role-boundary regression coverage, and production build pass.
- Threat model, retention schedule, executable retention cleanup, data-processing inventory, and CI security/integration jobs exist.
- Redis caches explicit session revocations while PostgreSQL remains the durable authority; replay detection still revokes the complete token family.
- Authenticated buyer and contributor profile editors use a server-side BFF and expose loading, success, and error states.
- Optional Alloy, Loki, and Grafana infrastructure centralizes container logs with a provisioned authentication dashboard and 30-day local retention.
- NestJS structured logging uses an operational-field allowlist with credential-leak regression coverage; Go emits the same core JSON schema.

The observability Compose profile validates successfully. Its first local runtime start on 7 September 2026 was blocked while Docker Desktop downloaded images by a Docker content-store `input/output error`; no application container or configuration error was reported. Retry after Docker Desktop storage/engine recovery.

On 8 September 2026, Docker Desktop recovered: core infrastructure became healthy, auth and web production images built successfully, their non-root/read-only runtime smoke checks passed, and the live media dependency integration suite passed. The optional observability profile has not yet been retried.

## Remaining release gates

These are intentionally not represented as completed in `MILESTONE_2_PLAN.md`:

- Browser automation and full assistive-technology verification.
- Deployment-platform scheduling of the tested retention command and legal/privacy approval.
- Production dashboards/alerts, live CI scan results, container vulnerability results, and privacy/legal sign-off.

Milestone 2 engineering functionality and local integration evidence are complete. Production release must still wait for the external operational, accessibility, and legal approval gates above.
