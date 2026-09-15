# Milestone 1 verification status

Last verified: 8 September 2026 (Africa/Lagos)

## Verified

- Root npm clean install completed.
- Identity service TypeScript checks, 12 Jest tests, and production build passed.
- Identity PostgreSQL migration ran successfully against local infrastructure.
- Identity OpenAPI export and validation passed.
- Media service formatting check, vet, unit/integration tests, build, migration, API startup, and worker compilation passed.
- Media API liveness, readiness, and catalogue endpoints returned HTTP 200.
- Identity API liveness and readiness endpoints returned HTTP 200.
- PostgreSQL, Redis, Kafka, and object-storage readiness checks passed before Docker became unresponsive.
- Kafka foundation topics existed and Redis returned `PONG`.
- Web TypeScript checks, ten Vitest tests, and the Next 16 production build passed.
- Public, authentication, buyer, and contributor route groups compile.
- Both OpenAPI contracts pass the root `contracts:validate` command.
- The production npm dependency audit reports zero vulnerabilities after upgrading Next.js and PostCSS.
- Compose configuration and the PostgreSQL initialization script validate syntactically.
- Application Dockerfiles declare non-root runtime users.
- After Docker Desktop recovery, PostgreSQL, Redis, Kafka, and MinIO pulled cleanly and all became healthy.
- Auth, media API, FFmpeg media worker, and web production images built successfully.
- Image inspection confirmed non-root runtime users for auth (`node`), media API (`nonroot:nonroot`), media worker (`65532:65532`), and web (`node`).
- Auth bcrypt and web Node runtime smoke checks passed with read-only filesystems.
- The non-root worker generated a synthetic 1920x1080 H.264 source, a 1280x720 H.264/AAC preview, and a JPEG poster with FFmpeg 6.1.2.
- Media migrations `001` through `003` applied and survived an infrastructure restart.
- Live PostgreSQL, Redis, Kafka, and MinIO Go integration tests passed against the recovered stack.

## Docker incident history

Docker Desktop stopped responding to its named-pipe API during the media image build. Repeated Bash commands bounded by a 20-second timeout returned no daemon response. Repository files and named volumes were not deleted or reset.

On 7 September 2026 the daemon became reachable again, but PostgreSQL, Redis, and MinIO simultaneously reported unhealthy, an observability image pull failed with a Docker content-store `input/output error`, and a bounded auth-service build failed when BuildKit returned `rpc error: code = Unavailable ... EOF`. No volumes or Docker data were deleted by verification. On 8 September 2026, the recovered engine completed the core infrastructure pulls, application builds, persistence check, and FFmpeg smoke test described above.

Regression commands:

```bash
docker-compose.exe -f tamo-infrastructure/docker-compose.yml up -d
docker-compose.exe -f tamo-infrastructure/docker-compose.yml ps
docker build -t tamo-auth-service:milestone2 tamo-auth-service
docker build -t tamo-media-api:milestone3 tamo-media-service
docker build -f tamo-media-service/Dockerfile.worker -t tamo-media-worker:milestone3 tamo-media-service
docker build -t tamo-web:milestone2 tamo-web
docker image inspect tamo-auth-service:milestone2 tamo-media-api:milestone3 tamo-media-worker:milestone3 tamo-web:milestone2 --format '{{.RepoTags}} user={{.Config.User}}'
```

Run the persistence test only after recording a known database row, Redis key, Kafka message, and MinIO object. Restart containers without `down -v`, then confirm each value remains.

## External repository controls

Default-branch protection, required GitHub checks, and the baseline commit require repository-owner decisions or credentials. They remain unchecked in the plan until configured in GitHub. Container vulnerability scanning remains pending now that all images build successfully.

## Milestone decision

The application, contract, infrastructure, image-build, non-root runtime, and media persistence foundations are verified. Milestone 1 is not formally closed until container vulnerability thresholds and external repository controls are verified.
