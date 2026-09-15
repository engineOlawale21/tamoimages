# Tamo Images platform contracts

These rules are normative for the web application, identity service, media API, background workers, and future services. They preserve independently deployable service boundaries inside the monorepo.

## Identity email events

`identity.email-verification-requested.v1` carries `userId`, `email`, and the one-time `verificationToken` to the authorized email-delivery consumer. `identity.password-recovery-requested.v1` carries `userId`, `email`, and the one-time `recoveryToken`. These payloads contain personal/security data: topics require restricted ACLs, transport encryption outside local development, no payload logging, bounded broker retention, and consumer-side deletion after delivery. Tokens are plaintext only in this delivery path; PostgreSQL stores their SHA-256 hashes.

Every event envelope includes a unique event ID, versioned event type, timestamp, and correlation ID. Consumers must be idempotent by event ID. Contract-breaking changes require a new topic version.

## HTTP conventions

- Public APIs use HTTPS and a major-version prefix such as `/api/v1`.
- Resource names are lowercase plural nouns. JSON fields use `camelCase`.
- Timestamps use UTC RFC 3339 with an explicit `Z` suffix.
- Identifiers are opaque strings; clients must not infer meaning from them.
- Collection responses use `items` and `page` metadata containing `cursor`, `nextCursor`, and `hasMore`. Cursor values are opaque.
- Mutation retries require an `Idempotency-Key` when the operation can create duplicate state.
- Errors use `application/problem+json` with `type`, `title`, `status`, `detail`, `instance`, and `correlationId`. Production responses never include stack traces or dependency details.

## Correlation IDs

The standard header is `X-Correlation-ID`. A trusted edge may generate it. Services accept only 1–128 characters from `[A-Za-z0-9._:-]`; invalid values are replaced. The value is returned in the response, included in structured logs, forwarded on internal HTTP calls, and copied into every Kafka event envelope. Correlation IDs are diagnostic metadata, not authentication credentials.

## Authentication contract

Bearer access tokens use asymmetric signing in production and include:

- `iss`: configured Tamo identity issuer
- `aud`: the intended service audience
- `sub`: opaque user ID
- `role`: `buyer`, `contributor`, or a separately governed administrative role
- `iat` and `exp`: issued-at and expiry timestamps
- `jti`: unique token identifier

Every service validates signature, algorithm, issuer, audience, expiry, and allowed role. Browser access tokens must not be stored in local storage. Refresh credentials use `Secure`, `HttpOnly`, and appropriate `SameSite` cookies and are rotated after use.

Local defaults may use `http://localhost:4000/api/v1` as issuer, `tamo-web` as browser audience, and service-specific audiences such as `tamo-media-service`. Staging and production values must be explicit configuration.

## Kafka event envelope

```json
{
  "id": "opaque-event-id",
  "type": "media.uploaded",
  "version": 1,
  "occurredAt": "2026-09-07T00:00:00Z",
  "correlationId": "request-or-workflow-id",
  "producer": "tamo-media-service",
  "payload": {}
}
```

The Kafka message key is the aggregate or event ID. Headers include `event-id`, `event-type`, and `correlation-id`. Producers own their event schemas. Consumers must tolerate additive fields. Removing or changing a field, its meaning, or its type requires a new event version and compatibility review. Dead-letter topics use `<source-topic>.dlq` and retain the original envelope plus failure classification; secrets and unnecessary personal data are prohibited.

| Topic | Producer | Consumers |
| --- | --- | --- |
| `media.uploaded.v1` | Media API | Media-processing workers |
| `media.processed.v1` | Media-processing workers | Media API/catalogue projections |
| `media.processing-failed.v1` | Media-processing workers | Media API, operations |
| `platform.smoke-test.v1` | Integration tests only | Integration tests only |

## Timeouts and retries

Browser-to-edge requests should complete within 30 seconds. Service HTTP connection timeouts are at most 3 seconds unless explicitly justified; ordinary internal request deadlines are at most 10 seconds. Health dependency checks target 750 milliseconds. Kafka and object-storage operations use bounded deadlines and retry only transient failures with exponential backoff and jitter.

Cancellation propagates through HTTP contexts and worker shutdown contexts. Non-idempotent operations are never retried without an idempotency mechanism. Services must stop accepting work during shutdown, finish or safely abandon admitted work within the configured grace period, then close database, Redis, Kafka, and object-storage clients.
