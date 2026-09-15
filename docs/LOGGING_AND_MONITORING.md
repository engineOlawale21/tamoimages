# Logging and monitoring standard

All application logs are newline-delimited JSON on standard output. Required fields are `timestamp`, `severity`, `service`, and `event`. HTTP completion events also contain `correlationId`, method, route template or path, status, and duration. They must never contain query strings, bodies, email addresses, names, passwords, password hashes, authorization headers, cookies, access/refresh tokens, verification/recovery tokens, object-storage credentials, or database URLs.

The NestJS logger enforces an allowlist of operational fields and has a credential-redaction regression test. The Go API emits the same core schema. Security audit records are separate, access-controlled PostgreSQL evidence; they use fixed event names and hashed subjects rather than raw identity values.

## Local centralized logging

Start the optional observability profile:

```bash
docker compose -f tamo-infrastructure/docker-compose.yml --profile observability up -d
```

Grafana is available at `http://localhost:3001`; Loki is available only for local administration at `http://localhost:3100`. Change the Grafana password in `.env` before use. Alloy discovers local Docker containers, parses JSON fields, and forwards them to Loki. The provisioned **Tamo authentication operations** dashboard includes service logs, HTTP failures, and security-event volume.

## Access, retention, and alerts

- Development Loki retention is 30 days. Production retention must use the privacy-approved class and storage lifecycle policy.
- Dashboard access is least-privilege and authenticated. Security/audit access is limited to approved security, privacy, and incident-response roles; access itself must be audited by the production provider.
- Alert on refresh replay, sustained authentication failures or 429 responses, unusual password-recovery volume, repeated 5xx responses, readiness failures, and retention-job failure.
- Alerts contain event counts, service, environment, status, and correlation IDs only. Operators retrieve restricted evidence through an approved incident process.
- Never turn on verbose framework, SQL-value, Kafka-payload, or HTTP-body logging in production.

Loki is an operational log store, not the authoritative security-audit database. Production alert destinations, escalation rotations, and retention approval remain environment-owned release controls.
