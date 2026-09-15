# Milestone 1 Plan: Runnable Foundation

## 1. Objective

Milestone 1 establishes a reliable engineering foundation for Tamo Images. At completion, every project must install, build, test, run, and deploy independently. The milestone does not deliver complete product functionality; it creates the platform on which authentication, uploads, processing, search, commerce, and contributor workflows can be implemented safely.

The four projects remain independent:

| Project | Technology | Local port |
| --- | --- | --- |
| `tamo-web` | Next.js | `3000` |
| `tamo-auth-service` | NestJS | `4000` |
| `tamo-media-service` | Go | `5000` |
| `tamo-infrastructure` | Docker Compose | Infrastructure-specific ports |

## 2. Milestone outcomes

Milestone 1 is complete when:

- One Git repository contains four independently deployable projects with path-aware release lifecycles.
- Each project can be configured without editing source code.
- Local infrastructure starts predictably.
- Both backend services expose liveness and readiness endpoints.
- The public APIs have valid OpenAPI documentation.
- The frontend can reach both backend services through typed API adapters.
- All projects produce structured, correlated logs.
- Unit, integration, formatting, linting, and build checks run in CI.
- Docker images run as non-root users where possible.
- A new developer can start the platform from the READMEs.

## 3. Scope

### Included

- Polyglot monorepo setup
- Dependency and toolchain locking
- Environment configuration and validation
- Local PostgreSQL, Redis, Kafka, and MinIO
- Database migration foundations
- Health and readiness checks
- Structured logging and correlation IDs
- OpenAPI validation
- Typed frontend API adapters
- Docker images
- CI pipelines
- Basic automated tests
- Developer documentation
- Security and operational defaults
- NDPR/NDP Act privacy-by-design backend baseline
- Production load-balancer/ingress requirements and stateless-service constraints

### Excluded

- Production-ready registration and login
- Email verification and password recovery
- Persistent media uploads
- FFmpeg processing
- Search implementation
- Cart, payment, licensing, and payouts
- Kubernetes manifests
- Full cloud infrastructure
- Complete implementation of supplied UI screens

## 4. Execution order

Work should proceed in the following order:

1. Clean and validate the development machine.
2. Establish the root monorepo and deployable boundaries.
3. Stabilize local infrastructure.
4. Stabilize the NestJS service.
5. Stabilize the Go service.
6. Stabilize the Next.js application.
7. Define and validate cross-service contracts.
8. Add CI pipelines.
9. Add local end-to-end smoke tests.
10. Complete documentation and milestone review.

Later sections may be developed in parallel after the infrastructure interfaces and API conventions are fixed.

---

## 5. Workstream A: development-machine readiness

### Goal

Ensure builds are reproducible and are not blocked by local disk, cache, runtime, or network problems.

### Tasks

- [x] Record available disk space.
- [x] Remove only incomplete `node_modules`, temporary build output, and disposable package caches.
- [x] Confirm Node.js and npm versions.
- [x] Confirm the Go toolchain version.
- [x] Install and confirm Docker Desktop or another compatible Docker engine.
- [ ] Confirm Git is installed and configured.
- [x] Confirm required ports are available.
- [x] Document Windows-specific setup considerations.
- [x] Decide whether runtime versions are pinned with `.nvmrc`, Volta, or documentation.
- [x] Pin the Go version in `go.mod` and CI.

### Recommended baseline

- Node.js: current active LTS
- npm: version bundled with the chosen Node.js release
- Go: one explicitly supported version
- Docker Compose: Compose v2
- PostgreSQL: supported stable release
- Redis: supported stable release
- Kafka: supported stable KRaft-based release

### Verification

```powershell
node --version
npm --version
go version
docker --version
docker compose version
git --version
```

### Acceptance criteria

- Package installation can complete without disk errors.
- Docker commands work from the terminal.
- Required ports are not occupied by unidentified processes.
- Supported runtime versions are documented.

---

## 6. Workstream B: polyglot monorepo setup

### Goal

Establish one polyglot monorepo while keeping every deployable independently testable, versionable, and releasable.

### Tasks

- [x] Add root npm workspace and orchestration commands.
- [x] Consolidate nested Git metadata into one root Git repository.
- [x] Create a project-specific `.gitignore`.
- [x] Add `.editorconfig`.
- [x] Add a `LICENSE` or document the private-license status.
- [ ] Add `CODEOWNERS` when team ownership is known.
- [x] Add a pull-request template.
- [x] Add issue templates if the repositories will be publicly collaborative.
- [x] Define branch protection expectations.
- [x] Define semantic versioning rules.
- [x] Add a changelog strategy.
- [x] Ensure generated output, secrets, IDE state, and dependencies are ignored.
- [ ] Make an initial baseline commit in the root repository.

### Repository rule

The repository has one root Git history and an npm workspace for the JavaScript projects. The Go module and infrastructure project retain their own manifests. Cross-service source imports remain prohibited: services integrate through versioned HTTP and Kafka contracts.

### Acceptance criteria

- Each deployable can be built and tested from its own directory.
- Root commands and CI verify every deployable.
- No deployable imports unpublished implementation source from another service directory.

---

## 7. Workstream C: infrastructure foundation

### Goal

Provide a predictable local runtime for PostgreSQL, Redis, Kafka, and object storage.

### Target services

| Service | Purpose | Port |
| --- | --- | --- |
| PostgreSQL | Persistent relational data | host `5433`, container `5432` |
| Redis | Sessions, caching, locks, rate limits | `6379` |
| Kafka | Asynchronous domain events | `9092` |
| MinIO API | S3-compatible object storage | `9000` |
| MinIO console | Local storage administration | `9001` |

### Tasks

- [x] Pin all container image versions; avoid `latest`.
- [x] Add health checks for every long-running container.
- [x] Add named volumes for persistent local state.
- [x] Add a shared Docker network.
- [x] Add resource limits suitable for local development.
- [x] Configure Kafka in KRaft mode.
- [x] Configure Kafka advertised listeners correctly for host and container clients.
- [x] Create separate PostgreSQL databases or schemas for identity and media ownership.
- [x] Add initialization scripts for databases.
- [x] Create the initial MinIO media bucket automatically.
- [x] Add a local bootstrap script for Kafka topics.
- [x] Add a reset script that clearly identifies destructive actions.
- [x] Add `.env.example` with safe development defaults.
- [x] Keep real credentials out of version control.
- [x] Document Windows Docker Desktop requirements.

### Kafka topics for foundation testing

- `platform.smoke-test.v1`
- `media.uploaded.v1`
- `media.processed.v1`
- `media.processing-failed.v1`

Production topic configuration will be finalized in later milestones. Foundation topics only need enough configuration to verify producers and consumers.

### Infrastructure checks

```powershell
docker compose config
docker compose up -d
docker compose ps
docker compose logs postgres
docker compose logs redis
docker compose logs kafka
docker compose logs minio
```

### Acceptance criteria

- `docker compose up -d` starts all services.
- Every service becomes healthy within the documented startup window.
- PostgreSQL accepts authenticated connections.
- Redis responds to a ping.
- Kafka accepts a produced message and a consumer receives it.
- MinIO accepts an object upload and download.
- Restarting containers preserves named-volume data.

---

## 8. Workstream D: NestJS identity-service foundation

### Goal

Create a production-shaped NestJS service foundation before implementing the complete identity domain.

### Target structure

```text
tamo-auth-service/
  src/
    main.ts
    app.module.ts
    config/
      configuration.ts
      environment.schema.ts
    common/
      errors/
      filters/
      interceptors/
      middleware/
    modules/
      auth/
      users/
      health/
    integrations/
      database/
      kafka/
      redis/
  test/
    integration/
    e2e/
  migrations/
  Dockerfile
  package.json
  package-lock.json
  nest-cli.json
  tsconfig.json
  tsconfig.build.json
  .env.example
  README.md
```

### Configuration tasks

- [x] Add a single typed configuration module.
- [x] Validate required variables during startup.
- [x] Reject weak or missing JWT secrets outside tests.
- [x] Validate port, URLs, token lifetimes, and environment names.
- [x] Add graceful shutdown hooks.
- [x] Configure CORS from an explicit allowlist.
- [x] Add a configurable API prefix and version.

### Database tasks

- [x] Select and document the PostgreSQL library or ORM.
- [x] Configure a bounded database connection pool.
- [x] Add the first migration framework configuration.
- [x] Create a migration metadata table.
- [x] Add a minimal users table migration or health-check-only migration.
- [x] Add migration run and rollback scripts.
- [x] Prevent automatic schema synchronization in production.

### Redis tasks

- [x] Add a Redis client provider.
- [x] Verify connection during readiness checks.
- [x] Namespace keys by environment and service.
- [x] Define connection timeout and retry behavior.
- [x] Close connections during graceful shutdown.

### Kafka tasks

- [x] Add a producer abstraction.
- [x] Configure client ID, brokers, retries, and timeouts.
- [ ] Add a smoke-test event publisher for integration tests.
- [x] Propagate event and correlation IDs.
- [x] Close the producer during shutdown.

### HTTP platform tasks

- [x] Add global validation with payload whitelisting.
- [x] Add a consistent problem-details exception filter.
- [x] Add request IDs and correlation IDs.
- [x] Add structured request logging.
- [x] Add request-size limits.
- [x] Configure security headers.
- [x] Add rate-limit infrastructure.
- [x] Ensure stack traces are hidden in production responses.
- [x] Redact credentials, tokens, cookies, private object URLs, and sensitive personal data from logs.
- [x] Attach a documented processing purpose and retention class to every persisted personal-data field introduced in this milestone.
- [x] Add an auditable data-subject-request interface boundary for later access, correction, export, and deletion workflows.
- [x] Document incident and personal-data-breach escalation ownership.

### Health endpoints

- `GET /api/v1/health/live`: confirms the process event loop is alive.
- `GET /api/v1/health/ready`: checks PostgreSQL, Redis, and required Kafka connectivity.

Readiness must fail if a required dependency cannot support requests.

### Swagger tasks

- [x] Document title, version, authentication scheme, and server URLs.
- [x] Add tags and descriptions.
- [x] Add explicit response DTOs.
- [x] Export OpenAPI JSON during CI.
- [x] Validate the exported specification.
- [x] Disable or protect interactive Swagger in production if required.

### Testing tasks

- [x] Unit-test configuration validation.
- [x] Test liveness without external dependencies.
- [x] Test readiness with healthy dependencies.
- [x] Test readiness failure behavior.
- [x] Test validation-error problem responses.
- [x] Add an application bootstrap smoke test.

### Build checks

```powershell
npm ci
npm run lint
npm test
npm run build
npm run start
```

### Acceptance criteria

- The service fails fast when configuration is invalid.
- The service builds with strict TypeScript settings.
- Swagger/OpenAPI generation succeeds.
- Liveness and readiness have distinct behavior.
- PostgreSQL, Redis, and Kafka integrations are exercised by tests.
- The runtime container uses a non-root user.

---

## 9. Workstream E: Go media-service foundation

### Goal

Create an operational Go API and worker foundation with clean dependency boundaries.

### Target structure

```text
tamo-media-service/
  cmd/
    api/
      main.go
    media-worker/
      main.go
  internal/
    config/
    health/
    media/
      domain/
      application/
      repository/
    platform/
      database/
      kafka/
      redis/
      storage/
      logging/
    transport/
      http/
      consumer/
  api/
    openapi.yaml
  migrations/
  testdata/
  Dockerfile
  go.mod
  go.sum
  .env.example
  README.md
```

### Configuration tasks

- [x] Create a typed configuration package.
- [x] Validate required configuration on startup.
- [x] Support environment variables without hidden defaults in production.
- [x] Configure HTTP timeouts explicitly.
- [x] Add graceful shutdown using signal-aware contexts.
- [x] Define worker concurrency and shutdown behavior.

### Database tasks

- [x] Select and document the PostgreSQL driver and query approach.
- [x] Configure connection-pool limits and timeouts.
- [x] Add migration tooling.
- [x] Add a minimal media schema migration.
- [x] Add transaction helpers without leaking database types into domain logic.

### Redis tasks

- [x] Add a Redis adapter.
- [x] Apply service and environment key prefixes.
- [x] Define timeout and retry behavior.
- [x] Include Redis in readiness checks.

### Kafka tasks

- [x] Add producer and consumer adapters.
- [x] Define consumer groups explicitly.
- [x] Add correlation and event IDs.
- [x] Handle graceful consumer shutdown.
- [x] Add retry classification.
- [x] Reserve a dead-letter topic convention.
- [x] Add one smoke-test consumer.

### Object-storage tasks

- [x] Add an S3-compatible storage interface.
- [x] Implement the MinIO adapter.
- [x] Verify bucket existence during bootstrap or provisioning.
- [x] Support put, get, head, delete, and presigned operations.
- [x] Add integration tests using a dedicated test bucket.

### HTTP platform tasks

- [x] Add versioned routing.
- [x] Add JSON content-type enforcement.
- [x] Add problem-details errors.
- [x] Add request and correlation IDs.
- [x] Add structured access logs.
- [x] Add panic recovery.
- [x] Add CORS configuration.
- [x] Add request-body and multipart limits.
- [x] Add secure server timeouts.

### Health endpoints

- `GET /health/live`: process liveness.
- `GET /health/ready`: PostgreSQL, Redis, Kafka, and object-storage readiness.

### OpenAPI tasks

- [x] Complete API metadata and server definitions.
- [x] Define reusable schemas and problem responses.
- [x] Document security requirements.
- [x] Validate the specification in CI.
- [ ] Decide whether handlers or clients will be generated from the contract.
- [ ] Detect breaking API changes during pull requests.

### Testing tasks

- [x] Unit-test configuration parsing.
- [x] Unit-test HTTP middleware.
- [x] Test liveness and readiness behavior.
- [x] Integration-test PostgreSQL.
- [x] Integration-test Redis.
- [x] Test Kafka producer-consumer flow.
- [x] Test MinIO object roundtrip.
- [x] Add race-detector execution in CI.

### Build checks

```powershell
go mod tidy
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

### Acceptance criteria

- Both API and worker binaries build independently.
- The API shuts down gracefully.
- Readiness reflects all required dependencies.
- Kafka and object-storage smoke tests pass.
- The race detector passes.
- The final container runs as a non-root user.

---

## 10. Workstream F: Next.js foundation

### Goal

Create a responsive, accessible frontend shell with stable API and authentication boundaries.

### Target structure

```text
tamo-web/
  src/
    app/
      (public)/
      (auth)/
      (buyer)/
      (contributor)/
    components/
      layout/
      ui/
    features/
      auth/
      media/
      uploads/
    lib/
      api/
      auth/
      config/
      telemetry/
      validation/
    styles/
  public/
  tests/
  Dockerfile
  next.config.ts
  package.json
  package-lock.json
  tsconfig.json
  .env.example
  README.md
```

### Configuration tasks

- [x] Validate public and server-only environment variables separately.
- [x] Add identity and media API base URLs.
- [x] Prevent server secrets from entering the client bundle.
- [x] Configure allowed image sources.
- [x] Add security headers.
- [x] Establish route groups for public, authentication, buyer, and contributor pages.

### Design-system foundation

- [x] Extract colors, spacing, typography, radii, and shadows from the supplied screens.
- [x] Define CSS variables or a token module.
- [x] Build accessible button variants.
- [x] Build text field, select, checkbox, radio, toggle, and textarea primitives.
- [x] Build modal, dropdown, toast, empty-state, card, and pagination primitives.
- [x] Build responsive header, footer, profile sidebar, and gallery layouts.
- [x] Define mobile, tablet, desktop, and wide-screen behavior.
- [x] Add focus, disabled, error, loading, and success states.

### API integration tasks

- [x] Add an identity API adapter.
- [x] Add a media API adapter.
- [x] Standardize timeouts and error mapping.
- [x] Forward correlation IDs when present.
- [ ] Generate clients from OpenAPI where practical.
- [x] Separate server-only and browser-safe clients.
- [x] Add mock adapters for component and end-to-end testing.

### Authentication boundary tasks

- [x] Define the session-cookie strategy.
- [x] Add protected route middleware scaffolding.
- [x] Add buyer and contributor authorization helpers.
- [x] Define unauthenticated, forbidden, and expired-session behavior.
- [x] Ensure access tokens are not stored in local storage.

### Quality tasks

- [x] Enable strict TypeScript checks.
- [ ] Configure ESLint and formatting. (TypeScript checking is active; ESLint/formatter policy remains.)
- [ ] Add component-test tooling.
- [ ] Add Playwright.
- [ ] Add accessibility checks.
- [x] Add a production bundle build in CI.
- [x] Define performance budgets for public pages.

### Initial test routes

- Public landing page
- Authentication placeholder page
- Buyer placeholder page
- Contributor placeholder page
- API connectivity/status page available only in development

### Build checks

```powershell
npm ci
npm run lint
npm run type-check
npm test
npm run build
npm run start
```

### Acceptance criteria

- Public, buyer, contributor, and authentication route groups compile.
- The shared shell is responsive and keyboard accessible.
- API errors are mapped to a stable frontend error model.
- No secret environment variable appears in the client output.
- The production build starts successfully in its container.

---

## 11. Workstream G: cross-service contracts

### Goal

Ensure independently deployable projects integrate through versioned contracts rather than shared implementation code.

### Tasks

- [x] Establish HTTP naming, pagination, timestamp, and error conventions.
- [x] Define a standard correlation-ID header.
- [x] Define authentication bearer-token claims.
- [x] Define service audience and issuer values.
- [x] Add reusable OpenAPI problem-response schemas.
- [x] Validate both OpenAPI documents.
- [ ] Generate frontend clients in an isolated CI step.
- [x] Define Kafka event-envelope schemas.
- [x] Add a schema compatibility rule.
- [x] Document event ownership and producers.
- [x] Document synchronous and asynchronous timeout expectations.

### Proposed standard headers

- `Authorization: Bearer <token>`
- `X-Request-ID`
- `X-Correlation-ID`
- `Idempotency-Key`

### Error response shape

```json
{
  "type": "https://api.tamoimages.com/problems/validation-error",
  "title": "Validation failed",
  "status": 422,
  "detail": "One or more fields are invalid.",
  "instance": "/api/v1/uploads",
  "requestId": "uuid",
  "errors": {
    "kind": ["kind must be image, video, or illustration"]
  }
}
```

### Acceptance criteria

- OpenAPI validation passes.
- Generated clients compile against the frontend.
- Both services use the same problem-response and correlation conventions.
- Event schemas are versioned and documented.

---

## 12. Workstream H: containerization

### Goal

Produce small, secure, repeatable application images.

### Tasks for all application images

- [x] Use multi-stage builds.
- [x] Pin base-image major versions or immutable digests for releases.
- [x] Copy only required files into build contexts.
- [x] Add `.dockerignore`.
- [x] Run the final image as a non-root user.
- [x] Avoid shipping build tools in runtime images.
- [x] Add OCI labels for version and source revision.
- [x] Add a container health check where appropriate.
- [ ] Scan images for known vulnerabilities.
- [ ] Test startup with read-only filesystem constraints where practical.

### Acceptance criteria

- Every application image builds from a clean checkout.
- Containers start with environment-only configuration.
- Containers do not require administrator/root permissions.
- Image vulnerability reports have no unresolved critical findings.

---

## 13. Workstream I: continuous integration

### Goal

Prevent code that cannot build, test, or satisfy contracts from reaching protected branches.

### NestJS pipeline

1. Checkout
2. Configure Node.js cache
3. `npm ci`
4. Lint
5. Type check
6. Unit tests with coverage
7. Integration tests with service containers
8. Build
9. Export and validate OpenAPI
10. Build and scan Docker image

### Go pipeline

1. Checkout
2. Configure Go cache
3. Verify modules
4. Formatting check
5. `go vet`
6. Unit tests
7. Race detector
8. Integration tests
9. Build API and worker
10. Validate OpenAPI
11. Build and scan Docker image

### Next.js pipeline

1. Checkout
2. Configure Node.js cache
3. `npm ci`
4. Lint
5. Type check
6. Unit and component tests
7. Accessibility tests
8. Production build
9. Playwright smoke tests
10. Build and scan Docker image

### Infrastructure pipeline

1. Validate Compose syntax
2. Validate environment template
3. Scan configuration for secrets
4. Start the stack
5. Wait for health checks
6. Run PostgreSQL, Redis, Kafka, and MinIO smoke tests
7. Shut down the stack

### CI policies

- [ ] Require all checks before merge.
- [x] Cancel outdated pipeline runs.
- [x] Cache dependencies without caching secrets.
- [ ] Upload test and coverage reports.
- [ ] Pin third-party CI actions.
- [x] Use least-privilege workflow permissions.
- [x] Do not publish images from pull-request builds.

### Acceptance criteria

- The default branch is protected.
- All projects have required status checks.
- A deliberately failing test blocks a merge.
- A clean branch produces deployable artifacts.

---

## 14. Workstream J: local integration and smoke tests

### Goal

Prove that the monorepo projects work together without coupling their service implementation trees.

### Smoke-test sequence

1. Start infrastructure.
2. Apply identity database migrations.
3. Apply media database migrations.
4. Start the NestJS service.
5. Start the Go API.
6. Start the Go worker.
7. Start the Next.js frontend.
8. Check liveness endpoints.
9. Check readiness endpoints.
10. Load both Swagger/OpenAPI documents.
11. Publish and consume a Kafka smoke-test event.
12. Upload and download a fixture through MinIO.
13. Load the frontend and call both backend health adapters.
14. Stop each application gracefully.

### Acceptance criteria

- A documented manual smoke test succeeds.
- An automated smoke script exits non-zero on any failed component.
- Logs from one request can be followed across service boundaries using its correlation ID.
- Stopping one dependency produces an accurate readiness failure.

---

## 15. Workstream K: documentation

### Goal

Allow a new developer to understand and run each project without private setup knowledge.

### Each project README must contain

- Purpose and ownership
- Technology and supported versions
- Prerequisites
- Environment-variable reference
- Local installation
- Development commands
- Test commands
- Build commands
- Docker commands
- Health endpoints
- API documentation location
- Database migration commands
- Troubleshooting section
- Deployment assumptions

### Platform documentation

- [x] Architecture overview
- [x] Local request-flow description
- [x] Port map
- [x] Service ownership matrix
- [x] API and Kafka conventions
- [x] Security assumptions
- [x] Data-ownership boundaries
- [x] Local reset and recovery procedure
- [x] Milestone definition of done

### Acceptance criteria

- A developer unfamiliar with the project can start it from documentation alone.
- All documented commands have been executed successfully.
- Environment templates describe every supported variable.

---

## 16. Security baseline

Milestone 1 must establish these controls even before full product functionality exists:

- [ ] No secrets committed to Git.
- [ ] Secret scanning in CI.
- [ ] Dependency vulnerability scanning.
- [ ] Container vulnerability scanning.
- [ ] Non-root runtime containers.
- [x] Explicit CORS allowlists.
- [x] Secure HTTP headers.
- [x] Request-size limits.
- [x] Server timeouts.
- [x] Redaction of authorization headers and credentials.
- [x] Production error responses without stack traces.
- [x] Database accounts with service-specific permissions.
- [ ] Separate development and test credentials.
- [x] Documented dependency-update process.

## 17. Observability baseline

### Required log fields

- Timestamp
- Severity
- Service name
- Service version
- Environment
- Request ID
- Correlation ID
- Route or operation
- Duration
- Status code or outcome
- Error code when applicable

### Tasks

- [x] Use JSON logs in backend services.
- [ ] Use human-readable pretty logs only as a local option.
- [x] Add HTTP request-duration measurements.
- [ ] Add dependency health duration and result logs.
- [x] Propagate correlation IDs into Kafka events.
- [ ] Add a placeholder OpenTelemetry configuration.
- [x] Document log redaction rules.

### Acceptance criteria

- One frontend-triggered request can be correlated across logs.
- Health checks do not flood normal logs.
- Sensitive headers and values are redacted.

## 18. Database migration policy

- Migrations are forward-only in deployed environments.
- Rollback files may exist for local use but production recovery must not rely on unsafe destructive reversal.
- Services apply only their own migrations.
- Application startup does not automatically mutate production schemas.
- CI verifies migrations on an empty database.
- CI verifies migrations against the previous release schema when possible.
- Schema changes must support rolling deployment compatibility.

## 19. Milestone risks and mitigations

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Insufficient disk space | Dependencies and builds fail | Clean partial dependencies and caches; document minimum space |
| Docker unavailable | Infrastructure cannot run | Install Docker Desktop and verify virtualization early |
| Kafka listener errors | Host or containers cannot connect | Test both advertised-listener paths |
| Unpinned dependencies | Non-reproducible builds | Commit lockfiles and pin runtime/container versions |
| Shared-database coupling | Services become difficult to extract | Enforce service-owned databases from the start |
| Contract drift | Frontend and services break independently | Validate OpenAPI and compile generated clients in CI |
| Over-engineering | Foundation delays feature delivery | Implement only interfaces needed for current dependencies |
| Secrets in local files | Credential exposure | Commit only examples and enable secret scanning |
| Readiness reports false success | Failed traffic after deployment | Check every request-critical dependency |

## 20. Suggested delivery sequence

### Sprint 1: machine and repositories

- Development-machine checks
- Disk-space recovery
- Independent Git repositories
- Lockfiles and runtime pinning
- Base documentation and ignore files

### Sprint 2: infrastructure

- Pinned Compose services
- Health checks and initialization
- Kafka topics and MinIO bucket
- Infrastructure smoke tests

### Sprint 3: backend foundations

- NestJS configuration, database, Redis, Kafka, health, logging, and Swagger
- Go configuration, database, Redis, Kafka, storage, health, logging, and OpenAPI

### Sprint 4: frontend and contracts

- Next.js route groups and UI primitives
- Typed API adapters
- Shared protocol conventions
- Generated-client validation

### Sprint 5: CI and integration

- Independent CI pipelines
- Container builds and scans
- Full local smoke test
- Documentation verification
- Milestone review

The actual duration depends on team size. A single engineer should treat these as sequential work packages rather than calendar-week commitments.

## 21. Definition of done

Milestone 1 is done only when all of the following are true:

### Infrastructure

- [x] Compose configuration validates.
- [x] PostgreSQL, Redis, Kafka, and MinIO become healthy.
- [x] Persistence survives restarts.
- [x] Infrastructure smoke tests pass.

### Identity service

- [x] Clean install succeeds.
- [x] Lint, type checks, tests, and build pass.
- [x] Database migrations run successfully.
- [x] Liveness and readiness behave correctly.
- [x] Swagger/OpenAPI validates.
- [ ] Container starts as non-root.

### Media service

- [ ] Module verification, formatting, vetting, tests, race checks, and builds pass. (Formatting, vet, tests, and builds pass locally; the race detector requires CGO and remains a Linux CI gate.)
- [x] API and worker binaries start.
- [x] Database migrations run successfully.
- [x] Liveness and readiness behave correctly.
- [x] Kafka and object-storage integration tests pass.
- [x] OpenAPI validates.
- [ ] Container starts as non-root.

### Frontend

- [x] Clean install succeeds.
- [x] Lint, type checks, tests, and production build pass.
- [x] Public and protected route foundations render.
- [ ] API adapters reach both services.
- [ ] Responsive and accessibility smoke tests pass.
- [ ] Container starts as non-root.

### Delivery process

- [x] All projects are tracked by one root Git repository.
- [ ] Required CI checks protect the default branch.
- [ ] Dependency and container scans pass the agreed threshold.
- [ ] A full local integration smoke test passes.
- [ ] Setup documentation has been followed successfully from a clean environment.

## 22. First actionable backlog

The recommended first implementation batch is:

1. Recover enough disk space for clean package installations.
2. Install and verify Docker.
3. Remove incomplete dependency directories created by failed installations.
4. Add project-specific ignore and editor configuration files.
5. Initialize the four independent Git repositories.
6. Pin infrastructure images and add health checks.
7. Create service-specific PostgreSQL databases.
8. Add configuration validation to NestJS and Go.
9. Implement separate liveness and readiness endpoints.
10. Add structured request logging and correlation IDs.
11. Validate both OpenAPI documents.
12. Add independent CI pipelines.
13. Run the full milestone smoke test.
