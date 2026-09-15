# Tamo Images platform

This repository is the polyglot monorepo for the Tamo Images platform. Each deployable keeps its own manifest, Dockerfile, environment contract, and release boundary while sharing root orchestration, CI, and engineering standards.

| Project | Responsibility |
| --- | --- |
| `tamo-web` | Next.js web application |
| `tamo-auth-service` | NestJS identity, sessions, roles, and public API documentation |
| `tamo-media-service` | Go API and Kafka workers for uploads and media processing |
| `tamo-infrastructure` | Local Kafka, Redis, PostgreSQL, and MinIO stack |

Start infrastructure first, then follow the README inside each service.

## Repository model

- One Git repository owns the platform source and planning documents.
- npm workspaces manage the NestJS and Next.js projects.
- The Go module remains independently compilable under `tamo-media-service`.
- Docker Compose remains under `tamo-infrastructure`.
- Services must communicate through HTTP or Kafka contracts; workspace source imports across service boundaries are prohibited.

## Root commands

```powershell
npm install
npm run infra:up
npm run verify
```

Run a development server with `npm run auth:dev` or `npm run web:dev`. Go commands are orchestrated from the root with `npm run media:test` and `npm run media:build`.

Cross-service HTTP, authentication, correlation, timeout, error, and Kafka rules are defined in [`docs/PLATFORM_CONTRACTS.md`](docs/PLATFORM_CONTRACTS.md).

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for platform flows, ports, ownership, data boundaries, recovery, and the Milestone 1 definition of done. Repository ownership and release rules are in [`docs/REPOSITORY_GOVERNANCE.md`](docs/REPOSITORY_GOVERNANCE.md).
