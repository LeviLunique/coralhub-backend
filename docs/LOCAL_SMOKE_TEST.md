# Local Smoke Test

This guide shows how to verify the current CoralHub backend bootstrap locally.

It covers:

- PostgreSQL
- MinIO
- API
- worker

If your local environment is not configured yet, start with [LOCAL_ENV_SETUP.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/LOCAL_ENV_SETUP.md).

## Prerequisites

Before starting, confirm:

- you are in the repository root
- Docker Desktop is running
- your local `.env` exists

Current local ports:

- PostgreSQL: `5433`
- MinIO API: `9000`
- MinIO console: `9001`
- API: `8080`

## 1. Start Local Dependencies

Run:

```bash
make compose-up
```

Then check container status:

```bash
docker compose ps
```

Expected result:

- `postgres` is `healthy`
- `minio` is `Up`
- `minio-create-bucket` exits with code `0`

## 2. Test PostgreSQL Availability

Check if PostgreSQL is accepting connections:

```bash
pg_isready -h localhost -p 5433 -U coralhub
```

Expected result:

```text
localhost:5433 - accepting connections
```

## 3. Connect to PostgreSQL

Connect with `psql`:

```bash
psql -h localhost -p 5433 -U coralhub -d coralhub
```

When prompted, use the password from `.env`:

```text
DB_PASSWORD
```

Inside `psql`, run:

```sql
\dt
```

Then run:

```sql
select id, slug, display_name, active from tenants;
```

Then run:

```sql
select * from tenant_configs;
```

Important:

- `\dt` is a `psql` meta-command, not SQL
- run `\dt` by itself on its own line and press Enter
- run each `select ...;` statement separately
- if you paste `\dt` together with SQL, `psql` can try to treat the SQL text as extra arguments to `\dt`

Expected result:

- `tenants` table exists
- `tenant_configs` table exists
- `tenants` contains `coral-jovem-asa-norte`

Exit:

```sql
\q
```

## 4. Test MinIO In The Browser

Open:

```text
http://localhost:9001
```

Login with the values from `.env`:

- `STORAGE_ACCESS_KEY`
- `STORAGE_SECRET_KEY`

Expected result:

- login succeeds
- bucket `coralhub-local` exists

## 5. Test MinIO With Docker CLI

If you want a CLI check instead of the browser, run:

```bash
docker run --rm --env-file .env --network coralhub-backend_default --entrypoint /bin/sh minio/mc \
  -c 'mc alias set local http://minio:9000 "$STORAGE_ACCESS_KEY" "$STORAGE_SECRET_KEY" && mc ls local && mc ls "local/$STORAGE_BUCKET"'
```

Expected result:

- the output includes the configured bucket from `.env`
- with the current local setup, that bucket is `coralhub-local`

Important:

- `docker run` does not automatically read the repository `.env` file
- `--env-file .env` is required so the temporary `minio/mc` container receives `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY`, and `STORAGE_BUCKET`
- `--entrypoint /bin/sh` is required because the `minio/mc` image uses `mc` as its default entrypoint

## 6. Start The API

Run:

```bash
make run-api
```

Expected result:

- the process starts
- no PostgreSQL connection error appears
- the server listens on `:8080`

## 7. Test API Health Endpoints

In another terminal, run:

```bash
curl http://127.0.0.1:8080/healthz
```

Expected response:

```json
{"service":"coralhub-api","status":"ok"}
```

Also test:

```bash
curl http://127.0.0.1:8080/api/v1/healthz
```

Expected response:

```json
{"service":"coralhub-api","status":"ok"}
```

## 8. Start The Worker

In another terminal, run:

```bash
make run-worker
```

Expected result:

- the process starts
- no PostgreSQL connection error appears
- the worker stays running

## 9. Stop Local Processes

To stop the API or worker:

```text
Ctrl+C
```

To stop Docker services:

```bash
make compose-down
```

## 10. Troubleshooting

### PostgreSQL connection fails

Check:

- Docker is running
- `docker compose ps` shows `postgres` as `healthy`
- `.env` contains:

```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=coralhub
DB_PASSWORD=<your local password from .env>
DB_NAME=coralhub
DB_SSL_MODE=disable
```

### MinIO login fails

Check:

- `minio` container is running
- `.env` values match the MinIO credentials in `docker-compose.yml`

### API does not start

Check:

- PostgreSQL is running
- your `.env` file exists
- port `8080` is free

## 11. Current Limitation

At this stage, the bootstrap only verifies infrastructure connectivity.

You can test:

- app startup
- worker startup
- PostgreSQL connectivity
- tenant seed data
- MinIO reachability

You cannot test yet:

- file upload through the API
- MinIO-backed application storage flows
- presigned download URLs

Those require later vertical slices to be implemented.
