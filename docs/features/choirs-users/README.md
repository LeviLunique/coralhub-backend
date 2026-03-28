# Choirs And Users Feature

This document explains the Stage 2 choirs and users slice implemented on branch:

- `feat/choirs-users`

Commit:

- `c0b6b00`

## Goal

Implement the first tenant-owned CRUD slice beyond tenant bootstrap.

This slice adds:

- `choirs` persistence and HTTP flows
- `users` persistence and HTTP flows
- tenant-aware request context resolution for protected module routes
- explicit SQL and repository tests against PostgreSQL

## Important Sequencing Note

This slice keeps the recommended Stage 2 endpoint shape from the implementation guide:

- `POST /api/v1/choirs`
- `GET /api/v1/choirs`
- `GET /api/v1/choirs/{choirID}`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{userID}`

However, the full auth baseline from Stage 3 still does not exist yet.

Because of that, tenant context for these routes is currently resolved from:

- `X-Tenant-Slug`

This is a deliberate narrow pull-forward of tenant context middleware so Stage 2 can validate:

- tenant-aware routing
- tenant-scoped SQL
- tenant-owned write and read flows

It is not the final authorization model.
Stage 3 should replace or harden this with authenticated actor and tenant context handling aligned with ADR 0004.

## HTTP Flow

For protected routes in this slice:

1. the router requires `X-Tenant-Slug`
2. middleware resolves the tenant through the tenant module
3. tenant ID and slug are placed in request context
4. handlers stay thin and read tenant context from request context
5. services validate business input
6. repositories execute explicit tenant-scoped SQL through `sqlc`

## File-By-File Explanation

### [000002_init_choirs_and_users.up.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/migrations/000002_init_choirs_and_users.up.sql)

Adds the new tenant-owned tables:

- `choirs`
- `users`

Important schema choices:

- both tables include `tenant_id`
- both tables enforce tenant-local uniqueness
- `choirs` uses unique `(tenant_id, name)`
- `users` uses unique `(tenant_id, email)`

### [choirs.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/choirs.sql)

Adds explicit SQL for:

- create choir
- get choir by ID within tenant scope
- list choirs within tenant scope

### [users.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/users.sql)

Adds explicit SQL for:

- create user
- get user by ID within tenant scope
- list users within tenant scope

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/choirs/model.go)

Defines the choir module model and create input.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/choirs/service.go)

Implements choir validation and application logic.

Current validations:

- tenant ID must be present
- choir name must be non-blank
- description is trimmed and optional

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/choirs/repository.go)

Defines the small repository contract used by the choir service.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/choirs/http.go)

Registers the choir routes and translates service errors into HTTP responses.

Current endpoints:

- `POST /api/v1/choirs`
- `GET /api/v1/choirs`
- `GET /api/v1/choirs/{choirID}`

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/users/model.go)

Defines the user module model and create input.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/users/service.go)

Implements user validation and application logic.

Current validations:

- tenant ID must be present
- email must parse as a valid address
- email is normalized to lowercase
- full name must be non-blank

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/users/repository.go)

Defines the focused repository contract used by the user service.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/users/http.go)

Registers the user routes and handles HTTP mapping.

Current endpoints:

- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{userID}`

### [middleware.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/http/middleware.go)

Adds `RequireTenantContext`.

What it does:

- reads `X-Tenant-Slug`
- resolves the active tenant using the tenant module
- rejects missing or unknown tenants
- injects tenant context into the request

### [context.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/requestctx/context.go)

Provides the shared request-context helpers for tenant data.

This avoids coupling business modules back to the platform HTTP package.

### [router.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/http/router.go)

Now wires:

- public tenant bootstrap routes
- protected choir routes
- protected user routes

The protected routes are grouped behind tenant context middleware.

### [main.go](/Users/levilunique/Workspace/Go/coralhub-backend/cmd/api/main.go)

Composes the new repositories and services into the running API.

### [choirs_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/choirs_repository.go)

Implements choir persistence with PostgreSQL and `sqlc`.

Important mappings:

- duplicate choir names become `ErrChoirNameTaken`
- unknown choir ID becomes `ErrChoirNotFound`
- every read and write is tenant-scoped

### [users_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/users_repository.go)

Implements user persistence with PostgreSQL and `sqlc`.

Important mappings:

- duplicate emails become `ErrUserEmailTaken`
- unknown user ID becomes `ErrUserNotFound`
- every read and write is tenant-scoped

### [repositories_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/repositories_integration_test.go)

Adds PostgreSQL-backed integration tests for choir and user repositories.

Why temporary tables are used:

- the tests exercise the real generated SQL
- the tests do not need to mutate the permanent application tables
- the tests still run against a real PostgreSQL connection

## Current Request Contract

Protected endpoints in this slice currently require:

```text
X-Tenant-Slug: coral-jovem-asa-norte
```

Example create choir request:

```http
POST /api/v1/choirs
X-Tenant-Slug: coral-jovem-asa-norte
Content-Type: application/json

{"name":"Sopranos","description":"Main choir"}
```

Example create user request:

```http
POST /api/v1/users
X-Tenant-Slug: coral-jovem-asa-norte
Content-Type: application/json

{"email":"ana@example.com","full_name":"Ana Clara"}
```

## What This Slice Does Not Yet Do

This slice does not yet implement:

- authenticated user identity
- role or membership checks
- update/delete flows for choirs and users
- device registration
- file upload
- event scheduling

Those remain in later stages of the roadmap.
