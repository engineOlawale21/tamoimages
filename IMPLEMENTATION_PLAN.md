# Tamo Images Monorepo Implementation Plan

## 1. Project boundaries

Tamo Images will be implemented as four deployable projects in one polyglot monorepo. JavaScript dependencies and commands are coordinated with npm workspaces, while Go and Docker Compose retain native manifests and tooling.

| Project | Responsibility |
| --- | --- |
| `tamo-web` | Next.js UI, routing, forms, rendering, and API integration |
| `tamo-auth-service` | NestJS authentication, sessions, roles, and account profiles |
| `tamo-media-service` | Go uploads, processing, media catalogue, batches, and collections |
| `tamo-infrastructure` | Kafka, Redis, PostgreSQL, MinIO, and local orchestration |

Billing, collections, checkout, and earnings will initially be implemented as clearly separated domains within the appropriate service. They should become standalone services only when scaling or ownership requirements justify the operational cost.

## 2. System architecture

### Communication

- The browser calls NestJS over HTTPS for identity and account operations.
- The browser calls Go over HTTPS for media, batch, collection, and catalogue operations.
- Services use Kafka for asynchronous workflows.
- Redis stores sessions, rate-limit counters, temporary upload state, caches, and locks.
- Each backend service owns its PostgreSQL schema or database.
- MinIO provides S3-compatible storage locally; production uses managed object storage.
- An API gateway or ingress exposes a single public origin.
- Production traffic passes through a managed load balancer or ingress with TLS termination and health-based routing. It routes identity and media paths to independently scalable service pools.

### Architectural rules

- A service may not access another service's database.
- HTTP is used for immediate request-response operations.
- Kafka is used for long-running work and cross-domain state changes.
- API contracts are distributed through OpenAPI-generated clients, not shared source packages.
- Large media files are uploaded directly from the browser to object storage.
- Every external operation that can be retried must be idempotent.
- Backend instances must remain stateless at the HTTP tier so a load balancer can distribute requests without sticky sessions. Temporary in-memory repositories are development-only and must be removed before horizontal scaling.

### Production edge and load balancing

The production request path is `CDN/WAF -> managed load balancer or ingress -> API routing layer -> auth/media service instances`. The selected cloud platform owns the implementation choice; examples include a managed application load balancer or Kubernetes ingress. Local development continues to use direct service ports and does not require a synthetic Nginx tier.

The production edge must:

- Terminate TLS using managed certificates and redirect HTTP to HTTPS.
- Route only to instances whose readiness probe succeeds while using liveness for restart decisions.
- Drain connections during deployments for at least the configured graceful-shutdown window.
- Preserve or generate the approved correlation header and replace untrusted forwarding headers.
- Apply request and connection limits, WAF controls, and denial-of-service protection before application traffic.
- Forward the original client network address only from trusted proxy hops; applications must not trust arbitrary public forwarding headers.
- Keep authentication and media scaling policies independent.
- Export availability, latency, error-rate, rejected-request, and healthy-target metrics.
- Avoid session affinity; sessions and rate-limit state belong in Redis or another shared store.

## 3. Phase one: foundations

The monorepo and each deployable require:

- One root Git history with path-aware ownership and releases
- `.gitignore` and `.editorconfig`
- Environment validation
- Dockerfile
- Root CI with project-specific jobs and path-aware checks
- Health and readiness endpoints
- Structured logging
- Unit and integration test configuration
- Root orchestration plus project-specific setup and operations documentation

Environment profiles must be provided for local development, automated tests, staging, and production. Production secrets must come from a secret manager.

### Acceptance criteria

- Every project installs, builds, tests, and starts independently.
- Infrastructure starts with Docker Compose.
- Health endpoints report dependency readiness.
- Cross-service source imports are prohibited; integration uses explicit contracts.

## 4. Phase two: contracts and domain modelling

### Core entities

- User
- Buyer profile
- Contributor profile
- Refresh session
- Media asset
- Media variant
- Batch and batch item
- Collection and collection item
- License
- Cart and order
- Download entitlement
- Earning and payout
- Release document

Use UUIDv7 or ULID identifiers. Identifiers exchanged between services are opaque strings.

### API conventions

- Prefix endpoints with `/api/v1`.
- Use cursor pagination for media feeds.
- Return RFC 9457-style problem responses.
- Include correlation IDs in requests, responses, logs, and events.
- Support idempotency keys for uploads, purchases, and payouts.
- Publish an OpenAPI contract for every public HTTP API.
- Version Kafka event schemas independently.

### Initial Kafka topics

- `identity.user-created.v1`
- `identity.profile-updated.v1`
- `media.upload-requested.v1`
- `media.uploaded.v1`
- `media.processing-started.v1`
- `media.processed.v1`
- `media.processing-failed.v1`
- `media.submitted-for-review.v1`
- `media.review-completed.v1`
- `commerce.order-paid.v1`
- `commerce.download-created.v1`
- `earnings.credit-created.v1`

Standard event envelope:

```json
{
  "eventId": "uuid",
  "eventType": "media.uploaded.v1",
  "occurredAt": "ISO-8601",
  "correlationId": "uuid",
  "producer": "tamo-media-service",
  "data": {}
}
```

## 5. Phase three: NestJS identity service

Replace the current in-memory account storage with PostgreSQL.

### Target structure

```text
src/
  main.ts
  app.module.ts
  config/
  common/
    decorators/
    exceptions/
    guards/
    interceptors/
    middleware/
  modules/
    auth/
      application/
      domain/
      infrastructure/
      presentation/
    users/
    profiles/
    sessions/
  integrations/
    kafka/
    redis/
  database/
    migrations/
```

### Delivery order

1. Buyer and contributor registration
2. Email verification
3. Login
4. Short-lived access tokens
5. Rotating refresh tokens
6. Redis-backed session management
7. Logout from one or all devices
8. Forgot/reset password
9. Role and permission guards
10. Profile management
11. Account suspension and deletion
12. Security audit events

### Security requirements

- Argon2id or bcrypt password hashing
- Refresh-token reuse detection
- Account- and IP-based rate limiting
- Generic authentication errors
- Secure HTTP-only refresh cookies
- Access-token issuer and audience validation
- No secrets or credentials in logs
- DTO validation on every endpoint
- CSRF protection for cookie-authenticated mutations

### Nigerian data-protection requirements (mandatory)

Both backend services must implement the Nigeria Data Protection Act 2023, the Nigeria Data Protection Regulation 2019 (NDPR), and applicable current Nigeria Data Protection Commission guidance as a mandatory security and privacy baseline. Where instruments differ, the current Act and binding NDPC guidance take precedence. Product launch requires review by the appointed privacy/legal owner; this engineering plan is not a substitute for legal advice.

- Maintain a data inventory and record of processing activities that identifies purpose, lawful basis, data category, recipients, storage location, retention period, and service owner.
- Collect only data necessary for specified, explicit purposes. New uses require compatibility and lawful-basis review.
- Publish versioned privacy notices and retain evidence of consent where consent is the lawful basis. Consent withdrawal must be as straightforward as giving consent.
- Provide authenticated workflows for access, correction, deletion, restriction, objection, portability, consent withdrawal, and complaints. Requests and fulfilment decisions must be audited without copying sensitive payloads into logs.
- Implement documented retention schedules and automated deletion or irreversible anonymisation for expired user, session, upload, media-metadata, support, and audit data. Legal holds must be explicit and auditable.
- Apply encryption in transit and at rest, least privilege, separation of service databases, secret management, secure password hashing, session revocation, dependency patching, backups, recovery testing, and privacy-safe observability.
- Never log passwords, tokens, session identifiers, raw authorization headers, payment data, government identifiers, precise addresses, or uploaded private media URLs. Pseudonymise user identifiers in operational telemetry where practical.
- Maintain processor agreements and due diligence for cloud, email, analytics, payments, storage, support, and observability providers. Record subprocessors and processing locations.
- Review cross-border transfers before enabling an overseas provider and document the applicable transfer basis and safeguards.
- Conduct and approve a Data Protection Impact Assessment before high-risk processing, including biometrics, large-scale profiling, sensitive data, automated decisions, or new AI-based media analysis.
- Maintain an incident register and a tested breach-response workflow that supports risk assessment, evidence preservation, NDPC notification, and notification to affected data subjects when legally required.
- Determine and document whether Tamo Images is a Data Controller or Data Processor of Major Importance, whether registration is required, whether a DPO must be appointed, and which compliance-audit filing obligations apply.
- Treat child data and sensitive personal data as restricted data requiring additional authorization, age/guardian controls where applicable, and a documented lawful basis.
- Add privacy and security acceptance tests to CI and retain evidence for periodic NDPR/NDP Act compliance audits.

Detailed engineering controls: [`BACKEND_SECURITY_NDPR_STANDARD.md`](./BACKEND_SECURITY_NDPR_STANDARD.md)

### Documentation and tests

- Complete Swagger descriptions and response schemas
- Authentication-rule unit tests
- PostgreSQL integration tests
- Redis session tests
- Registration, login, refresh, and logout end-to-end tests

## 6. Phase four: Go media service

Use ports-and-adapters boundaries while keeping domain code straightforward.

### Target structure

```text
cmd/
  api/
  media-worker/
internal/
  media/
    domain/
    application/
    repository/
  batch/
  collection/
  catalogue/
  upload/
  processing/
  platform/
    database/
    kafka/
    redis/
    storage/
    observability/
  transport/
    http/
    consumer/
api/
  openapi.yaml
migrations/
```

### Upload workflow

1. The frontend requests an upload session.
2. Go validates the media type and contributor permissions.
3. Go creates a pending asset record.
4. Go returns a presigned multipart storage URL.
5. The browser uploads directly to object storage.
6. The browser confirms upload completion.
7. Go publishes `media.uploaded.v1`.
8. A worker scans and inspects the file.
9. The worker extracts metadata.
10. The worker produces thumbnails or transcodes video.
11. Generated variants are stored.
12. The worker publishes a success or failure event.
13. The frontend receives status through polling initially, followed by SSE if required.

### Processing capabilities

- MIME signature validation
- File-size and resolution validation
- Malware scanning
- EXIF extraction and sanitization
- Image resizing and WebP/AVIF previews
- Video inspection with FFprobe
- HLS transcoding with FFmpeg
- Poster-frame extraction
- Retryable processing jobs
- Dead-letter handling
- Idempotent Kafka consumers

### Domain delivery order

1. Direct uploads
2. Processing status
3. Contributor media libraries
4. Batches and batch items
5. Media metadata and releases
6. Review submission
7. Collections and favourites
8. Public catalogue
9. Search and filtering
10. Licensed downloads

## 7. Phase five: Next.js frontend

### Target structure

```text
src/
  app/
    (public)/
    (auth)/
    (buyer)/
    (contributor)/
  features/
    auth/
    uploads/
    media/
    batches/
    collections/
    checkout/
    profile/
  components/
    ui/
    layout/
  lib/
    api/
    auth/
    validation/
    telemetry/
  styles/
```

Business components belong in feature folders. Only reusable, domain-neutral primitives belong in `components/ui`.

### UI delivery order

1. Shared design tokens, header, footer, buttons, fields, cards, and modals
2. Responsive public landing page
3. Registration and login
4. Search results and filters
5. Image and video detail pages
6. Contributor dashboard shell
7. Photo, video, and illustration libraries
8. Upload interface with progress and retry states
9. Batch creation and submission
10. Collections and favourites
11. Buyer profile and download history
12. Pricing, cart, and checkout
13. Contributor earnings and payouts
14. Custom-content and case-study pages

### Frontend standards

- React Server Components by default
- Client components only for required interaction
- Generated API clients
- React Hook Form and Zod for forms
- Keyboard-accessible components
- Mobile-first responsive layouts
- Loading, empty, error, and retry states
- Image optimization and paginated galleries
- No authentication tokens in local storage
- Route-level error boundaries

## 8. Phase six: search, commerce, and licensing

### Search

Begin with indexed PostgreSQL search. Add OpenSearch only when relevance or scale requirements demonstrate the need.

Supported filters:

- Keywords and tags
- Media type
- Creative or editorial
- Orientation
- Resolution
- Contributor
- Location
- Newest or relevance ordering

### Commerce

Implement:

- Cart management
- License selection
- Server-calculated pricing
- Payment-provider checkout
- Verified payment webhooks
- Immutable order records
- Download entitlements
- Expiring signed download URLs
- Invoices and receipts

The frontend must never be trusted as the source of prices or entitlements.

### Contributor earnings

- Ledger-based earning records
- Pending and available balances
- Reversals instead of destructive edits
- Payout requests and administrative review
- Idempotent payment callbacks
- Complete financial audit trail

## 9. Phase seven: observability and reliability

### Observability

- Structured JSON logs
- OpenTelemetry traces
- Prometheus metrics
- Correlation IDs across HTTP and Kafka
- Central error reporting
- Kafka consumer-lag monitoring
- Upload and processing-duration metrics
- Failed-job dashboards
- Storage-usage alerts
- Database-pool monitoring

### Reliability patterns

- Transactional outbox for database-to-Kafka publishing
- Consumer inbox for event deduplication
- Exponential retry with jitter
- Dead-letter topics
- Graceful process shutdown
- Liveness and readiness probes
- Controlled database migration steps

## 10. Testing strategy

### Next.js

- Component tests
- API adapter tests
- Accessibility tests
- Playwright tests for critical journeys
- Responsive visual-regression tests

### NestJS

- Application-service unit tests
- PostgreSQL and Redis integration tests
- OpenAPI contract tests
- Authentication security tests

### Go

- Table-driven domain tests
- Repository integration tests
- Multipart upload tests
- Kafka consumer tests
- FFmpeg processing fixtures
- Idempotency and failure tests

### Critical end-to-end journeys

1. A contributor registers and verifies an account.
2. The contributor uploads a photo or video.
3. Media processing completes.
4. The contributor creates a batch and submits it.
5. A reviewer approves the asset.
6. A buyer discovers and purchases the asset.
7. The buyer downloads the licensed file.
8. The contributor receives an earnings credit.

## 11. Delivery milestones

### Milestone 1: runnable foundation

- All monorepo projects build successfully through root and project-local commands.
- Infrastructure starts locally.
- CI pipelines pass.
- Health checks and API documentation work.

Detailed execution plan: [`MILESTONE_1_PLAN.md`](./MILESTONE_1_PLAN.md)

### Milestone 2: authentication

- Registration, email verification, login, token refresh, and logout
- Buyer and contributor profiles
- Protected Next.js routes

Detailed execution plan: [`MILESTONE_2_PLAN.md`](./MILESTONE_2_PLAN.md)

### Milestone 3: media ingestion

- Direct multipart uploads
- Kafka processing pipeline
- Image and video variants
- Progress, failure, and retry UI

### Milestone 4: contributor workflow

- Contributor media libraries
- Batches
- Metadata and release forms
- Review submission and status

### Milestone 5: buyer discovery

- Search and filters
- Media detail pages
- Favourites and collections

### Milestone 6: commerce

- Pricing and licensing
- Cart and checkout
- Purchased downloads
- Receipts

### Milestone 7: production readiness

- Earnings ledger and payouts
- Observability
- Security review
- Load testing
- Backup and recovery exercises

## 12. Immediate implementation backlog

1. Free disk space and remove incomplete dependency installations.
2. Consolidate projects into one root Git repository and npm workspace.
3. Lock dependency versions and generate lockfiles.
4. Add PostgreSQL migrations to the identity service.
5. Replace in-memory authentication storage.
6. Add Redis-backed refresh sessions.
7. Add Kafka, PostgreSQL, Redis, and S3 adapters to Go.
8. Finalize and validate the OpenAPI contracts.
9. Generate the Next.js API clients.
10. Complete registration and login end to end.
11. Implement the presigned multipart upload workflow.
12. Add the first media-processing worker.
