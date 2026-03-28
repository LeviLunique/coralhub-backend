# Memberships And Authorization Baseline Testing Guide

This file shows how to verify the Stage 3 slice on branch:

- `feat/memberships-auth`

## 1. Confirm You Are On The Correct Branch

Run:

```bash
git branch --show-current
```

Expected:

```text
feat/memberships-auth
```

## 2. Start Local Dependencies

Run:

```bash
make compose-up
```

## 3. Apply The Stage 2 And Stage 3 Migrations

Run:

```bash
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000002_init_choirs_and_users.up.sql
PGPASSWORD='<your DB_PASSWORD from .env>' psql -h localhost -p 5433 -U coralhub -d coralhub -f db/migrations/000003_init_choir_members.up.sql
```

## 4. Start The API

Run:

```bash
make run-api
```

## 5. Create Two Users

Run:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","full_name":"Ana Clara"}' \
  http://127.0.0.1:8080/api/v1/users
```

Then:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"maria@example.com","full_name":"Maria Luz"}' \
  http://127.0.0.1:8080/api/v1/users
```

## 6. Create A Choir As The First Actor

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sopranos"}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- HTTP `201`
- the actor becomes the first choir `manager` automatically

Save the returned choir `id`.

## 7. Verify Choir Listing Is Membership-Scoped

As the manager:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana@example.com' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- the created choir appears

As a non-member:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria@example.com' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- empty `items`

## 8. Add A Membership

Use the saved choir ID and the `user_id` for `maria@example.com`.

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<maria-user-id>","role":"member"}' \
  http://127.0.0.1:8080/api/v1/choirs/<choir-id>/memberships
```

Expected:

- HTTP `201`

## 9. Verify The New Member Can See The Choir

Run:

```bash
curl -s \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: maria@example.com' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- the choir now appears

## 10. Verify Missing Actor Header Fails

Run:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected:

- HTTP `400`
- error about `X-User-Email`

## 11. Run Automated Checks

Run:

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go vet ./...
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go build ./cmd/api ./cmd/worker
```

Expected:

- all commands succeed
