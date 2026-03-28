# Tenant Bootstrap Testing Guide

This file shows how to verify the tenant bootstrap feature on branch:

- `feat/tenant-bootstrap`

Endpoint under test:

```text
GET /api/v1/public/tenants/{tenantSlug}
```

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/tenant-bootstrap
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

## 4. Start The API

Run:

```bash
make run-api
```

Leave it running in that terminal.

## 5. Call The New Endpoint

In another terminal, run:

```bash
curl -s http://127.0.0.1:8080/api/v1/public/tenants/coral-jovem-asa-norte
```

Expected response:

```json
{"slug":"coral-jovem-asa-norte","display_name":"Coral Jovem Asa Norte","branding":{}}
```

Note:

- `branding` is currently empty because `tenant_configs` has no seeded row yet

## 6. Test Unknown Tenant

Run:

```bash
curl -s -i http://127.0.0.1:8080/api/v1/public/tenants/tenant-that-does-not-exist
```

Expected:

- HTTP status `404`
- response body:

```json
{"error":"tenant not found"}
```

## 7. Test Blank Slug Behavior

Because the slug is a path segment, the easiest invalid request check is a trailing slash route miss:

```bash
curl -s -i http://127.0.0.1:8080/api/v1/public/tenants/
```

Expected:

- route not found behavior from `chi`
- typically `404`

The service-level blank-slug validation is still covered by unit tests.

## 8. Run Automated Checks

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

## 9. Stop Local Processes

Stop the API:

```text
Ctrl+C
```

Stop local dependencies if you want:

```bash
make compose-down
```

## 10. Troubleshooting

### The endpoint returns `404` for the seeded tenant

Check:

- you are on branch `feat/tenant-bootstrap`
- the API was rebuilt or restarted after switching branches
- PostgreSQL contains the seeded tenant row

### The API does not start

Check:

- `postgres` is healthy
- `.env` exists
- `DB_PASSWORD` matches the currently initialized local PostgreSQL volume

If you changed credentials after the volume already existed:

```bash
docker compose down -v
make compose-up
```

### The response has empty branding

That is expected right now.

There is no seeded `tenant_configs` row yet.
