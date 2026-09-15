# Milestone 2 plan: authentication, sessions, and profiles

## 1. Outcome

Deliver production-shaped account authentication for buyers and contributors, with verified email ownership, short-lived access tokens, rotating refresh sessions, protected frontend routes, profile management, auditable security events, and Nigerian data-protection controls.

Milestone 2 is complete when registration, verification, login, refresh, logout, recovery, authorization, and profile journeys pass automated integration and end-to-end tests.

## 2. Mandatory constraints

- PostgreSQL is the identity system of record.
- Redis stores bounded session and abuse-prevention state; it is not the sole record of account ownership.
- Passwords use bcrypt cost 12 or an approved Argon2id configuration.
- Access tokens are short-lived and validate issuer, audience, subject, role, token ID, and expiry.
- Refresh credentials are opaque, random, hashed at rest, rotated on every use, and reuse-detected by token family.
- Browser refresh credentials use `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`, and no `Domain` in production.
- Cookie-authenticated mutations enforce origin and CSRF controls.
- Authentication errors do not enable account enumeration.
- Operational logs never contain passwords, bearer tokens, refresh tokens, session cookies, verification/recovery tokens, or raw personal profiles.
- The Nigeria Data Protection Act 2023, NDPR 2019, applicable NDPC guidance, and `BACKEND_SECURITY_NDPR_STANDARD.md` are acceptance requirements.

## 3. Delivery order

### Workstream A: identity schema

- [x] Add account state and email-verification fields.
- [x] Add one-time verification-token storage using token hashes.
- [x] Add password-recovery-token storage using token hashes.
- [x] Add refresh-session families with rotation and revocation metadata.
- [x] Add buyer and contributor profile tables with data-minimised fields.
- [x] Add privacy-safe security audit events.
- [x] Add retention-oriented expiry indexes.
- [x] Apply and rollback the migration against a clean integration database.

### Workstream B: registration and email verification

- [x] Register buyer and contributor accounts as unverified.
- [x] Prevent registration from issuing access credentials.
- [x] Generate cryptographically random, single-use verification tokens.
- [x] Store only token hashes and bounded expiry timestamps.
- [x] Publish or enqueue a versioned verification-email request.
- [x] Verify email and activate the account transactionally.
- [x] Resend verification with account/IP rate limiting.
- [x] Return enumeration-resistant responses.
- [x] Add unit, integration, and HTTP tests.

### Workstream C: login and access tokens

- [x] Authenticate only active, verified, non-suspended accounts.
- [x] Add account- and IP-scoped throttling with bounded Redis expiry.
- [x] Configure JWT issuer, audience, token ID, and short expiry.
- [x] Add bearer strategy, authenticated-request type, and role guard foundation.
- [x] Record privacy-safe login success/failure audit events.
- [x] Add unit, integration, and HTTP tests.

### Workstream D: rotating sessions

- [x] Issue an opaque refresh credential and persist only its hash.
- [x] Rotate the refresh credential on every refresh.
- [x] Detect reuse and revoke the complete token family.
- [x] Support logout from the current session.
- [x] Support logout from all sessions.
- [x] Enforce absolute and idle expiries.
- [x] Bind useful device metadata without invasive fingerprinting.
- [x] Add Redis-backed revocation/cache behavior with PostgreSQL durability.
- [x] Add concurrency and replay tests.

### Workstream E: password recovery

- [x] Add enumeration-resistant forgot-password requests.
- [x] Generate expiring, single-use recovery tokens and store only hashes.
- [x] Reset the password transactionally.
- [x] Revoke all existing sessions after reset.
- [x] Record recovery security events without token or password data.
- [x] Add unit, integration, and HTTP tests.

### Workstream F: profiles and account lifecycle

- [x] Add authenticated buyer profile retrieval and update.
- [x] Add authenticated contributor profile retrieval and update.
- [x] Validate and minimise profile fields.
- [x] Add account suspension enforcement.
- [x] Add soft-deletion request and session revocation.
- [x] Connect access, correction, deletion, restriction, objection, portability, consent-withdrawal, and complaint requests to the compliance port.
- [x] Define retention defaults and legal-hold behavior. (Scheduled cleanup worker remains.)
- [ ] Add authorization and privacy tests.

### Workstream G: Next.js session boundary

- [x] Add server-side login, registration, verification, refresh, and logout adapters.
- [x] Issue the opaque `__Host-tamo-session` cookie only from the server boundary.
- [x] Add same-origin return-path validation.
- [x] Protect buyer and contributor layouts by validated session and role.
- [x] Add authenticated profile screens and navigation states.
- [x] Add expired-session, forbidden, loading, empty, success, and error states.
- [ ] Add accessible form validation and focus management.
- [ ] Add component and browser journey tests.

### Workstream H: contracts, operations, and security

- [x] Document implemented endpoints in generated OpenAPI.
- [x] Version Kafka/email event contracts and document personal-data fields.
- [x] Add verification, recovery, session, and audit retention periods to the processing inventory.
- [x] Add secret, dependency, and filesystem/container-configuration scans to CI.
- [x] Add authentication dashboards and privacy-safe alert fields.
- [x] Run threat modelling for enumeration, credential stuffing, fixation, replay, CSRF, XSS, and privilege escalation.
- [ ] Complete privacy/legal review before production enablement.

## 4. API target

| Method | Route | Purpose |
| --- | --- | --- |
| `POST` | `/api/v1/auth/register` | Create an unverified account |
| `POST` | `/api/v1/auth/verify-email` | Consume a verification token |
| `POST` | `/api/v1/auth/resend-verification` | Request a replacement token |
| `POST` | `/api/v1/auth/login` | Create a refresh session and access token |
| `POST` | `/api/v1/auth/refresh` | Rotate the refresh session |
| `POST` | `/api/v1/auth/logout` | Revoke the current session |
| `POST` | `/api/v1/auth/logout-all` | Revoke every account session |
| `POST` | `/api/v1/auth/forgot-password` | Request password recovery |
| `POST` | `/api/v1/auth/reset-password` | Consume recovery token and revoke sessions |
| `GET/PATCH` | `/api/v1/profiles/me` | Read or update the role-specific profile |

## 5. Initial retention defaults

These are engineering defaults pending the privacy owner’s approval:

- Email verification token: 30 minutes; consumed/expired records deleted within 24 hours.
- Password recovery token: 15 minutes; consumed/expired records deleted within 24 hours.
- Refresh session: 30-day absolute maximum and a shorter configurable idle timeout.
- Security audit event: 12 months unless incident response or legal hold requires documented extension.
- Unverified account: deletion after 7 days if verification is never completed.

## 6. Acceptance gates

- [x] Authentication lifecycle integration journeys pass through real PostgreSQL and Redis.
- [x] Refresh-token replay revokes the complete session family.
- [x] No plaintext credential or one-time token is stored or logged.
- [x] Buyer and contributor roles cannot cross protected boundaries.
- [x] Account deletion and password reset revoke active sessions.
- [x] OpenAPI and event contracts validate locally and are configured in CI.
- [x] Backend session cookies, replay behavior, and mutation origin protections pass HTTP integration tests.
- [ ] Data inventory, retention schedule, threat model, and audit evidence are complete.
- [ ] Root verification, container builds, vulnerability scans, and end-to-end tests pass.
