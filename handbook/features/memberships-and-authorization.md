# Memberships and Authorization

This guide explains the authorization baseline introduced after the initial choir and user slice.

## What It Does

This feature adds:

- choir memberships
- manager and member roles
- actor-aware protected flows
- membership-scoped choir visibility

Representative endpoint:

- `POST /api/v1/choirs/{choirID}/memberships`

## How It Works

Flow:

1. middleware resolves tenant context and actor identity
2. the actor is loaded inside the current tenant
3. services enforce membership and role checks
4. choir reads become membership-scoped instead of tenant-wide
5. choir creation auto-creates the first manager membership

The current actor resolution still uses request headers as a staged baseline, but the authorization model is explicitly tenant-aware.

## Why It Matters

This slice stops protected operations from relying on tenant context alone.
It establishes the first real role checks in the backend.

## How To Verify

Start dependencies and the API:

```bash
make compose-up
make run-api
```

Create two users, then create a choir as the first actor:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","full_name":"Ana Clara"}' \
  http://127.0.0.1:8080/api/v1/users
```

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' -H 'Content-Type: application/json' \
  -d '{"email":"maria@example.com","full_name":"Maria Luz"}' \
  http://127.0.0.1:8080/api/v1/users
```

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'X-User-Email: ana@example.com' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sopranos"}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Expected result:

- the choir is created
- the creator becomes the first manager

Then add a membership and verify visibility:

- the manager can add members
- a non-member cannot see the choir until added

Automated validation:

```bash
make test
make vet
make build
```
