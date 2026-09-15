# Milestone 2 authentication threat model

Reviewed scope: registration, verification, login, recovery, refresh sessions, profiles, privacy requests, and the Next.js backend-for-frontend boundary.

| Threat | Control | Evidence / verification |
| --- | --- | --- |
| Account enumeration | Generic registration, resend, recovery, and login responses; unknown-user login performs a bcrypt comparison | Auth service tests and HTTP response tests |
| Credential stuffing | IP and HMAC-pseudonymised account throttles with bounded Redis expiry | Rate-limit middleware tests and Redis integration test |
| Password disclosure | bcrypt cost 12; DTO bounds; body/header redaction; credentials absent from events | Unit tests and log review |
| Verification/recovery token theft | 256-bit random opaque tokens, SHA-256 hashes at rest, short expiry, single use | Repository integration tests |
| Session fixation | Server creates a new opaque refresh token on login and rotates on every refresh | Refresh integration test |
| Refresh replay | Row locking detects reuse and revokes the complete token family | Concurrent replay integration test |
| CSRF | Refresh/logout cookies are SameSite=Lax and mutation requests require the configured exact Origin | Middleware HTTP tests |
| XSS token theft | Refresh token is HttpOnly and access tokens remain at the server boundary | Cookie assertions and frontend review |
| Privilege escalation | Signed role claim, active-account lookup on refresh, role guards, role-specific profile tables | Authorization tests |
| Suspended/deleted account access | Login requires active state; refresh rechecks current state and revokes sessions; deletion request revokes all sessions | Account-lifecycle tests |
| Privacy-request impersonation | Privacy endpoints require a valid bearer subject; request records contain no free-form payload | Controller tests and schema review |
| Sensitive telemetry | Audit events allow only fixed event/outcome fields and hashed subjects; arbitrary metadata is not accepted | Audit service review |

Residual risks requiring operational ownership include email-provider compromise, leaked signing secrets, denial of service above the application tier, and administrator misuse. Production release requires secret rotation, alert ownership, dependency/container scan review, backup restoration evidence, and privacy-owner approval.
