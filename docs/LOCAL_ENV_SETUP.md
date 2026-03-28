# Local Environment Setup

This guide shows how to configure the CoralHub backend local environment before running the smoke test.

It covers:

- `.env` creation
- local password generation
- PostgreSQL and MinIO credential alignment
- local dependency startup

After completing this guide, continue with [LOCAL_SMOKE_TEST.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/LOCAL_SMOKE_TEST.md).

## 1. Prerequisites

Before starting, confirm:

- you are in the repository root
- Docker Desktop is running
- `openssl` is available
- `make` is available

Optional but useful:

- `psql`
- `pg_isready`

## 2. Create Your Local `.env`

Copy the example file:

```bash
cp .env.example .env
```

## 3. Generate A Password For `DB_PASSWORD`

Use a hex password to avoid shell and `.env` escaping problems:

```bash
openssl rand -hex 32
```

Example output:

```text
9f72d5f8f4b7d1b6d4e1a1d7c3b6a4d57f1d63e3d1c3e6a9b4d2c1f6a8e0d7c2
```

Copy the generated value and set it in `.env`:

```env
DB_PASSWORD=<generated value>
```

## 4. Generate A Password For `STORAGE_SECRET_KEY`

Run:

```bash
openssl rand -hex 32
```

Copy the generated value and set it in `.env`:

```env
STORAGE_SECRET_KEY=<generated value>
```

## 5. Edit `.env`

Update `.env` so it looks like this:

```env
APP_ENV=development
HTTP_ADDR=:8080
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s

DB_HOST=localhost
DB_PORT=5433
DB_USER=coralhub
DB_PASSWORD=<generated value>
DB_NAME=coralhub
DB_SSL_MODE=disable
DB_MAX_CONNS=10
DB_MIN_CONNS=1
DB_MAX_CONN_LIFETIME=30m
DB_MAX_CONN_IDLE_TIME=5m
DB_HEALTH_CHECK_PERIOD=30s

WORKER_POLL_INTERVAL=5s

STORAGE_ENDPOINT=localhost:9000
STORAGE_BUCKET=coralhub-local
STORAGE_REGION=us-east-1
STORAGE_ACCESS_KEY=coralhub
STORAGE_SECRET_KEY=<generated value>
STORAGE_USE_SSL=false

OTEL_SERVICE_NAME=coralhub-backend
LOG_LEVEL=INFO
```

Notes:

- `docker-compose.yml` reads PostgreSQL credentials from `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `DB_PORT`
- `docker-compose.yml` reads MinIO credentials from `STORAGE_ACCESS_KEY` and `STORAGE_SECRET_KEY`
- `STORAGE_BUCKET` is used by the MinIO bootstrap container

## 6. If You Changed Credentials, Reset Local Volumes

This is important.

If you changed `DB_PASSWORD`, `STORAGE_ACCESS_KEY`, or `STORAGE_SECRET_KEY` after containers or volumes already existed, recreate the local volumes:

```bash
docker compose down -v
```

Reason:

- PostgreSQL applies the initial credentials when the data directory is first created
- MinIO local state can also retain prior configuration in the existing volume

If this is your first local setup, this step is still safe.

## 7. Start Local Dependencies

Run:

```bash
make compose-up
```

Then check status:

```bash
docker compose ps
```

Expected result:

- `postgres` is `healthy`
- `minio` is `Up`
- `minio-create-bucket` exits with code `0`

## 8. Continue With The Smoke Test

Once the dependencies are up, continue with:

- [LOCAL_SMOKE_TEST.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/LOCAL_SMOKE_TEST.md)

## 9. Optional One-Liner Password Generation

If you want to generate both secrets quickly, run:

```bash
openssl rand -hex 32
openssl rand -hex 32
```

Use:

- the first value for `DB_PASSWORD`
- the second value for `STORAGE_SECRET_KEY`

## 10. Troubleshooting

### PostgreSQL password does not work

Usually this means:

- you changed `DB_PASSWORD` in `.env`
- but did not recreate the PostgreSQL volume

Fix:

```bash
docker compose down -v
make compose-up
```

### MinIO login does not work

Check:

- `.env` values for `STORAGE_ACCESS_KEY` and `STORAGE_SECRET_KEY`
- you recreated volumes if you changed the MinIO credentials

### Step 5 of `LOCAL_SMOKE_TEST.md` fails

Check:

- `.env` exists
- `docker run` uses `--env-file .env`
- the local MinIO container is running
