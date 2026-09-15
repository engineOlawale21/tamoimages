# Tamo media service

Go owns upload orchestration, metadata extraction, thumbnails, image optimization, and video transcoding. Milestone 3 begins with PostgreSQL-backed pending assets and direct presigned uploads to S3-compatible storage.

Copy `.env.example` to `.env` and start the independent infrastructure project before running `go run ./cmd/api`.

Apply PostgreSQL migrations explicitly with `go run ./cmd/migrate`. The service uses `pgx` with a bounded pool and transaction helper; it never performs automatic schema synchronization. The Redis adapter namespaces every key and applies bounded operation timeouts and retries. Kafka uses explicit consumer groups, manual commits, correlation/event headers, graceful context cancellation, retry classification, and the `<topic>.dlq` dead-letter convention. The S3-compatible adapter supports bucket checks, put/get/head/delete, and presigned upload/download URLs.

## Foundation endpoints

- `GET /api/v1/health/live` checks the API process only.
- `GET /api/v1/health/ready` checks PostgreSQL, Redis, Kafka, and S3-compatible storage reachability.
- `GET /api/v1/media` returns the temporary public catalogue.
- `POST /api/v1/upload-sessions` creates a contributor-owned pending asset and returns a short-lived presigned `PUT` URL.
- `POST /api/v1/upload-sessions/{id}/complete` verifies the stored object and marks the asset uploaded.
- `GET /api/v1/media/{id}` returns the current status only when the authenticated contributor owns the asset.

The upload session endpoint accepts metadata only. File bytes travel directly from the browser to object storage and never pass through the API. The media service verifies access tokens with the same HS256 secret, issuer, and audience configured by the identity service. Production deployments should inject that shared secret through a secret manager.

`go run ./cmd/media-worker` consumes `media.uploaded.v1` with a bounded consumer group. It atomically claims uploaded records, downloads each private object to an ephemeral directory, inspects it with FFprobe, creates derivatives with FFmpeg, uploads them, and records `ready` or a bounded privacy-safe failure code. Images receive WebP thumbnail (320px), small (640px), medium (1280px), and large (2048px) bounding-box variants. Videos receive JPEG thumbnail/poster variants and H.264/AAC MP4 small (480p), medium (720p), and large (1080p) bounding-box renditions. Processing never upscales, skips duplicate dimensions, and removes partially uploaded derivatives after a failure. The original object is never modified. Duplicate upload events are ignored while processing is active or after completion; a processing claim older than five minutes can be reclaimed after a crashed worker. Deterministic variant keys make that retry idempotent. Successful and failed processing publish `media.processed.v1` and `media.processing-failed.v1` respectively. Use `Dockerfile.worker` for the FFmpeg-enabled worker image; local execution requires `ffmpeg` and `ffprobe` on `PATH` or explicit `FFMPEG_PATH` and `FFPROBE_PATH` values.

`GET /api/v1/media/{id}` returns processing metadata and, once ready, ordered variant descriptors with short-lived signed download URLs. Private object-storage keys are never returned to clients.

The API contract is in `api/openapi.yaml`. HTTP responses include correlation IDs, secure headers, and problem details. Structured access logs deliberately exclude bodies, authorization data, object URLs, and user data under the project NDPR security standard.

## Verification

```powershell
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

Set `DATABASE_INTEGRATION_URL`, `REDIS_INTEGRATION_ADDR`, `KAFKA_INTEGRATION_BROKER`, `S3_INTEGRATION_ENDPOINT`, `S3_INTEGRATION_ACCESS_KEY`, and `S3_INTEGRATION_SECRET_KEY` to include the real dependency round-trip tests. The object-storage test creates and removes a dedicated bucket. CI provisions isolated dependencies for these tests.

Linux CI additionally executes the race detector and builds the non-root container image.

Persisted fields and object metadata must be reviewed against `docs/data-processing-inventory.md` and the repository-level NDPR backend standard before release.
