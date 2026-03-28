# Choirs And Users Testing Guide

This file shows how to verify the Stage 2 choirs and users slice on branch:

- `feat/choirs-users`

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/choirs-users
```

## 2. Start Local Dependencies

Run:

```bash
make compose-up
```

Then confirm:

```bash
docker compose ps
```

Expected:

- `postgres` is `healthy`
- `minio` is `Up`

## 3. Confirm The Seeded Tenant Exists

Run:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -c "select slug, display_name, active from tenants;"
```

Expected:

- one row for `coral-jovem-asa-norte`
- display name `Coral Jovem Asa Norte`
- `active = t`

## 4. Apply The New Migration To Your Local Database

This repository still does not have a migration runner wired into the app.
So for now, apply the Stage 2 migration manually:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000002_init_choirs_and_users.up.sql
```

Then confirm:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -c "\dt"
```

Expected:

- `choirs` exists
- `users` exists

## 5. Start The API

Run:

```bash
make run-api
```

Leave it running in that terminal.

## 6. Create A Choir

In another terminal, run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sopranos","description":"Main choir"}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- HTTP status `201`
- JSON body containing:
  - a generated `id`
  - `name = "Sopranos"`
  - `tenant_id` for the seeded tenant

## 7. List Choirs

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- JSON object with `items`
- one item with `name = "Sopranos"`

## 8. Create A User

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","full_name":"Ana Clara"}' \
  http://127.0.0.1:8080/api/v1/users
```

Expected:

- HTTP status `201`
- JSON body containing:
  - a generated `id`
  - `email = "ana@example.com"`
  - `full_name = "Ana Clara"`

## 9. List Users

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  http://127.0.0.1:8080/api/v1/users
```

Expected:

- JSON object with `items`
- one item with `email = "ana@example.com"`

## 10. Test Missing Tenant Context

Run:

```bash
curl -s -i http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- HTTP status `400`
- response body:

```json
{"error":"X-Tenant-Slug header is required"}
```

## 11. Run Automated Checks

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
```

Then:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
```

Then:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all commands complete successfully

## 12. Stop Local Processes

Stop the API:

```text
Ctrl+C
```

Stop local dependencies if you want:

```bash
make compose-down
```

## 13. Troubleshooting

### Requests return `400` with missing tenant header

Check:

- the request includes `X-Tenant-Slug`
- the value is `coral-jovem-asa-norte`

### Requests return `404 tenant not found`

Check:

- the seeded tenant row exists
- the header value matches the seeded tenant slug exactly

### The API starts but choir/user endpoints fail with relation errors

Check:

- you applied `db/migrations/000002_init_choirs_and_users.up.sql`
- you restarted the API after switching branches

### `go test ./...` skips or fails integration tests

Check:

- PostgreSQL is running locally
- `.env` points to the correct DB credentials
- the seeded tenant still exists in the local DB
