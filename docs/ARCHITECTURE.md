# Tamo Images architecture

## Platform overview

Tamo Images is a polyglot monorepo containing independently deployable applications. The Next.js web application calls versioned identity and media APIs. PostgreSQL databases are owned by one backend each, Redis holds ephemeral coordination state, Kafka carries asynchronous domain events, and S3-compatible object storage holds media bytes.

Production requests follow `CDN/WAF -> managed load balancer or ingress -> API route -> stateless service instance`. Local development uses direct host ports and does not require a local load balancer.

## Request flows

### Identity request

1. The browser sends a bounded HTTPS request with or receives an `X-Correlation-ID`.
2. The edge routes `/api/v1/auth/*` and identity account routes to the NestJS identity service.
3. The service validates input, applies Redis-backed rate limits, and reads or writes only the identity database.
4. It returns JSON or `application/problem+json`; logs contain metadata and correlation IDs, not credentials or personal payloads.

### Media upload request

1. The browser asks the Go media API for an upload operation.
2. The API authenticates the user, records bounded metadata in the media database, and returns a short-lived presigned object-storage URL.
3. The browser uploads bytes directly to object storage.
4. The API finalizes the upload and publishes `media.uploaded.v1`.
5. A bounded worker consumes the event, processes the object, and publishes `media.processed.v1` or `media.processing-failed.v1`.

### Health flow

Liveness confirms only that a process can serve requests. Readiness checks required dependencies using short deadlines. The load balancer sends traffic only to ready instances and uses liveness failures for restart decisions.

## Local port map

| Component | Host port | Container/internal port |
| --- | ---: | ---: |
| Next.js web | `3000` | `3000` |
| Identity API | `4000` | `4000` |
| Media API | `5000` | `5000` |
| PostgreSQL | `5433` | `5432` |
| Redis | `6379` | `6379` |
| Kafka external listener | `9092` | `9092` |
| Kafka internal listener | n/a | `29092` |
| MinIO API | `9000` | `9000` |
| MinIO console | `9001` | `9001` |

PostgreSQL uses host port `5433` because a native Windows PostgreSQL instance may already own `5432`. Applications inside the Compose network use container port `5432`.

## Ownership matrix

| Area | Owning project | Persistent data |
| --- | --- | --- |
| Browser UI and rendering | `tamo-web` | None; approved cookies only |
| Accounts, credentials, roles, sessions | `tamo-auth-service` | `tamo_identity` |
| Media metadata, uploads, processing state | `tamo-media-service` | `tamo_media`, object storage |
| Local dependency lifecycle | `tamo-infrastructure` | Docker named volumes |
| HTTP and Kafka conventions | Repository root | Versioned contract documents |
| NDPR engineering controls | Repository root and service owners | Processing inventories and restricted audit systems |

No service may query another service's database. Cross-domain communication uses documented HTTP APIs or versioned Kafka events.

## Security assumptions

- Production TLS, WAF controls, denial-of-service protection, and trusted forwarding-header handling are supplied by the managed edge.
- HTTP services remain stateless; session and rate-limit state belongs in Redis or another shared store.
- Production credentials come from a secret manager and are never committed or copied into logs.
- Access tokens are short lived and refresh credentials use secure HttpOnly cookies.
- Object-storage buckets are private; access uses narrowly scoped, short-lived presigned URLs.
- Backend processing follows `BACKEND_SECURITY_NDPR_STANDARD.md`, including minimisation, purpose limitation, retention, access control, auditability, and data-subject handling.

## Data boundaries

The identity service is authoritative for user identity and credential data. The media service stores only the user identifier needed to establish media ownership and must not duplicate profile or credential fields. Kafka payloads contain the minimum fields required by consumers. Raw media belongs in object storage; relational databases store object keys and controlled metadata rather than private URLs or bytes.

Every new personal-data field requires a documented purpose, lawful-basis owner, retention class, access policy, correction path, and deletion behavior in the owning service's processing inventory.

## Local reset and recovery

1. Inspect state with `docker compose --file tamo-infrastructure/docker-compose.yml ps --all` and bounded logs before changing anything.
2. Restart an individual unhealthy service first; restarting must preserve named volumes.
3. Re-run one-shot Kafka and MinIO bootstrap jobs when only provisioning is missing.
4. Apply service migrations explicitly after PostgreSQL is healthy.
5. Use `tamo-infrastructure/scripts/reset.ps1` only when local data loss is intended. It requires confirmation unless `-Force` is supplied.
6. Never add `--volumes` to routine shutdown commands. Back up required local data before a reset.

## Milestone 1 definition of done

Milestone 1 is complete only when every required checklist item in `MILESTONE_1_PLAN.md` is implemented and its acceptance evidence has passed from a clean environment. At minimum:

- dependency installation, formatting, linting, tests, race checks, builds, and validated OpenAPI contracts pass;
- Compose dependencies become healthy, retain data through restart, and pass PostgreSQL, Redis, Kafka, and MinIO round trips;
- migrations run explicitly and both APIs demonstrate distinct liveness and readiness behavior;
- final containers run as non-root and pass the agreed vulnerability threshold;
- root CI enforces the required checks without publishing pull-request images;
- a browser-to-service and asynchronous smoke test passes with propagated correlation IDs;
- security and NDPR controls, ownership, reset, and onboarding documentation have been exercised.
