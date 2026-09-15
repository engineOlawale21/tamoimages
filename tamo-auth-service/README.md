# Tamo identity service

NestJS owns registration, login, account roles, tokens, and sessions. The current in-memory repository is an explicit development adapter and must be replaced with PostgreSQL before production.

Configuration is validated before startup, request payloads are allowlisted, errors use problem details, and structured request logs carry correlation IDs without logging bodies, credentials, tokens, or cookies.

Registration and login are protected by Redis-backed rate limits. Client network identifiers are HMAC-pseudonymised before they become Redis keys and expire with the configured rate-limit window.

## Local setup

Copy `.env.example` to `.env`, start the independent infrastructure project, and run:

```powershell
npm ci
npm run start:dev
```

Apply or roll back PostgreSQL migrations explicitly:

```powershell
npm run migration:up
npm run migration:down
```

The service uses the `pg` driver with parameterized SQL and a bounded connection pool. Schema synchronization is never performed automatically. Redis keys are namespaced by environment and service. Kafka publishing uses versioned topics and an envelope containing event, correlation, producer, and occurrence identifiers.

## Endpoints

- `GET /api/v1/health/live` checks process liveness without external dependencies.
- `GET /api/v1/health/ready` checks PostgreSQL, Redis, and Kafka reachability and returns `503` if a dependency is unavailable.
- Swagger is exposed at `http://localhost:4000/docs` outside production.

## Verification

```powershell
npm run lint
npm test
npm run build
npm run openapi:export
```

Backend changes must follow [`BACKEND_SECURITY_NDPR_STANDARD.md`](../BACKEND_SECURITY_NDPR_STANDARD.md). Never persist a new personal-data field without documenting its purpose, lawful-basis owner, retention class, and deletion path.

The NDPR data-subject workflow boundary is defined in `src/compliance/data-subject-request.port.ts`. Incident escalation and breach-response ownership are documented in `docs/personal-data-breach-response.md`; named contacts must be maintained in the private incident-management system.
