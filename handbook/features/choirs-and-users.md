# Choirs and Users

This guide explains the first tenant-owned CRUD slice and how to verify it.

## What It Does

This feature adds tenant-scoped choir and user creation and read flows.

Representative endpoints:

- `POST /api/v1/choirs`
- `GET /api/v1/choirs`
- `GET /api/v1/choirs/{choirID}`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{userID}`

## How It Works

Flow:

1. protected routes require tenant context
2. middleware resolves the tenant from the request
3. handlers decode the payload and call the service layer
4. services validate input
5. repositories execute explicit tenant-scoped SQL

The core rule is that both `choirs` and `users` are tenant-owned from the start.

## Why It Matters

This slice establishes:

- tenant-aware CRUD structure
- explicit SQL discipline
- thin handlers with focused services

## How To Verify

Start dependencies and the API:

```bash
make compose-up
make run-api
```

Create a choir:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sopranos","description":"Main choir"}' \
  http://127.0.0.1:8080/api/v1/choirs
```

Create a user:

```bash
curl -s -i \
  -H 'X-Tenant-Slug: coral-jovem-asa-norte' \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","full_name":"Ana Clara"}' \
  http://127.0.0.1:8080/api/v1/users
```

List choirs:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' http://127.0.0.1:8080/api/v1/choirs
```

List users:

```bash
curl -s -H 'X-Tenant-Slug: coral-jovem-asa-norte' http://127.0.0.1:8080/api/v1/users
```

Expected result:

- created records are returned with tenant-scoped IDs
- list endpoints show only data within the current tenant

Automated validation:

```bash
make test
make vet
make build
```
