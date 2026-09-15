# Tamo infrastructure

Local dependencies for the independently deployed Tamo Images services.

## Prerequisites

- Docker Desktop with Compose v2
- At least 8 GB of free memory recommended for the complete stack

On Windows, restart the terminal after installing Docker Desktop so its CLI is added to `PATH`. If it is not found yet, the default per-user binary directory is `%LOCALAPPDATA%\Programs\DockerDesktop\resources\bin`.

## Start

Copy `.env.example` to `.env`, then run:

```powershell
docker compose config
docker compose up -d
docker compose ps
```

## Services

| Service | Host address |
| --- | --- |
| PostgreSQL | `localhost:5433` (container port `5432`) |
| Redis | `localhost:6379` |
| Kafka | `localhost:9092` |
| MinIO API | `http://localhost:9000` |
| MinIO console | `http://localhost:9001` |

Kafka containers use `kafka:29092` when communicating inside the `tamo-local` network. Host applications use `localhost:9092`.

Application databases are kept separate: `tamo_identity` and `tamo_media`. One-shot bootstrap containers create both databases, the foundation Kafka topics, and the configured MinIO bucket. Re-running `docker compose up -d` is safe.

Verify the bootstrap jobs completed successfully:

```powershell
docker compose ps -a
docker compose logs kafka-init
docker compose logs minio-init
```

## Stop

```powershell
docker compose down
```

Do not add `--volumes` unless local PostgreSQL, Redis, and MinIO data should be permanently deleted.

## Reset local data

The reset script prints a destructive-action warning and requests confirmation before deleting named volumes:

```powershell
.\scripts\reset.ps1
```

Use `-Force` only in an automated disposable environment where data loss is intended.
# Local Docker disk discipline

All infrastructure containers rotate their local JSON logs at 10 MB with three files retained per container. Run `npm run infra:trim` after image-development sessions to cap reusable build cache at 1 GB and remove dangling images. The command never prunes named volumes, so PostgreSQL, Redis, MinIO, Loki, Alloy, and Grafana data are preserved.
