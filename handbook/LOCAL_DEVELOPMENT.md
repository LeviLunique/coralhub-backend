# Local Development

This guide explains how to run CoralHub Backend locally and how to validate the current baseline before opening a pull request.

## Prerequisites

Minimum local requirements:

- Go
- Docker with Compose support
- `make`

Useful local tools:

- `sqlc`
- `psql`
- `pg_isready`

For the full local quality bar, you also need access to the tools used by `make ci`.

## Local Services

The local stack uses:

- PostgreSQL on host port `5433`
- MinIO API on port `9000`
- MinIO console on port `9001`
- API on port `8080`

## 1. Create `.env`

Copy the example file:

```bash
cp .env.example .env
```

The default example values are suitable for a first local run.

If you need stronger local secrets, update:

- `DB_PASSWORD`
- `STORAGE_SECRET_KEY`

## 2. Start Dependencies

Run:

```bash
make compose-up
```

This starts:

- PostgreSQL
- MinIO
- the bucket bootstrap helper

To verify container state:

```bash
docker compose ps
```

## 3. Apply Migrations

The repository stores migrations in `db/migrations/` using `golang-migrate` file naming.

If you already use a migration tool locally, you can use it here.

If you want a direct repository-native path, apply the migrations in order with `psql`:

```bash
for f in db/migrations/*.up.sql; do
  PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -f "$f"
done
```

Then confirm the seeded tenant exists:

```bash
PGPASSWORD="$DB_PASSWORD" psql \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  -c "select slug, display_name, active from tenants;"
```

Expected result:

- one row for `coral-jovem-asa-norte`
- display name `Coral Jovem Asa Norte`

## 4. Start the API

Run:

```bash
make run-api
```

Health checks:

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/api/v1/healthz
```

Expected response:

```json
{"service":"coralhub-api","status":"ok"}
```

## 5. Start the Worker

In a separate terminal:

```bash
make run-worker
```

The worker should stay running and connect to the same PostgreSQL database as the API.

## 6. Verify Storage

To inspect MinIO in the browser, open:

```text
http://localhost:9001
```

Use the credentials from `.env`:

- `STORAGE_ACCESS_KEY`
- `STORAGE_SECRET_KEY`

Expected result:

- login succeeds
- bucket `coralhub-local` exists

## 7. Useful Commands

Local development:

- `make compose-up`
- `make compose-down`
- `make run-api`
- `make run-worker`

Code quality:

- `make fmt`
- `make fmt-check`
- `make vet`
- `make staticcheck`
- `make lint`
- `make govulncheck`
- `make test`
- `make build`
- `make sqlc`
- `make ci`

`make ci` is the closest local equivalent to the Stage 11 CI baseline.

## 8. Environment Variables You Will Touch Most Often

HTTP:

- `HTTP_ADDR`
- `HTTP_HANDLER_TIMEOUT`

Database:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSL_MODE`

Worker:

- `WORKER_POLL_INTERVAL`
- `WORKER_MAX_ATTEMPTS`
- `WORKER_RETRY_BACKOFF`
- `WORKER_NOTIFICATION_RETENTION`

Storage:

- `STORAGE_ENDPOINT`
- `STORAGE_BUCKET`
- `STORAGE_ACCESS_KEY`
- `STORAGE_SECRET_KEY`
- `STORAGE_USE_SSL`

Push:

- `FIREBASE_ENABLED`
- `FIREBASE_CREDENTIALS_FILE`

## 9. Common Issues

### PostgreSQL credentials stopped working

If you changed database credentials after the volume was already created, recreate the local volumes:

```bash
docker compose down -v
make compose-up
```

### MinIO credentials no longer match

If you changed `STORAGE_ACCESS_KEY` or `STORAGE_SECRET_KEY`, recreate the local volumes the same way:

```bash
docker compose down -v
make compose-up
```

### `make ci` fails on `sqlc`

Install `sqlc` locally or run the command in CI.

### Push notifications are not sending locally

This is expected unless Firebase is enabled and a valid service account file is configured. The worker can still run locally with push delivery disabled.

### Migrations fail with missing credentials

Check that:

- `.env` exists
- the PostgreSQL container is running
- your shell has the same `DB_*` values as `.env`

If you prefer not to export the variables into your shell, replace the command with explicit values from `.env`.
