# Memberships And Authorization Baseline

This document explains the Stage 3 memberships and authorization baseline implemented on branch:

- `feat/memberships-auth`

Commit:

- `d433c15`

## Goal

Tighten the Stage 2 tenant-only protected flows by introducing:

- `choir_members`
- explicit actor request context
- membership-aware choir access
- first role-aware authorization checks

## What Changed

Stage 2 proved tenant-scoped CRUD structure, but choir routes still trusted only tenant context.

Stage 3 adds a narrow authorization baseline:

- choir routes now require both:
  - `X-Tenant-Slug`
  - `X-User-Email`
- actor user is resolved inside the current tenant
- choir reads are membership-scoped
- choir creation automatically creates the actor as the first choir `manager`
- choir membership management requires an existing `manager` membership

User routes remain under tenant-only context for now so the codebase still has a lightweight user bootstrap path.
That is a pragmatic staging choice, not the final auth model.

## Important Sequencing Note

This is still not a real production identity provider integration.

`X-User-Email` is an authorization stub for the current stage.
It exists to satisfy ADR 0004’s requirement that protected flows stop relying on tenant hints alone.

The long-term replacement should be authenticated user claims from a real auth boundary.

## New Runtime Flow

For actor-protected routes:

1. middleware reads `X-Tenant-Slug`
2. middleware resolves the active tenant
3. middleware reads `X-User-Email`
4. middleware resolves the actor user inside that tenant
5. tenant and actor are added to request context
6. handlers stay thin and call services
7. services and repositories enforce membership or role checks

## File-By-File Explanation

### [000003_init_choir_members.up.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/migrations/000003_init_choir_members.up.sql)

Adds the `choir_members` table.

Important schema choices:

- includes `tenant_id`
- includes `role`
- constrains role to:
  - `manager`
  - `member`
- enforces tenant-local uniqueness for `(tenant_id, choir_id, user_id)`

### [memberships.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/memberships.sql)

Adds explicit SQL for:

- create choir membership
- get actor membership in a choir
- list choir members with joined user identity fields

### [choirs.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/choirs.sql)

Adds member-scoped choir reads:

- get choir only if actor is a member
- list choirs only for the actor’s memberships

### [users.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/users.sql)

Adds tenant-scoped actor lookup by email:

- `GetUserByEmail`

### [context.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/requestctx/context.go)

Now stores both:

- tenant context
- actor context

### [middleware.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/http/middleware.go)

Adds `RequireActorContext`.

What it does:

- resolves tenant from `X-Tenant-Slug`
- resolves actor from `X-User-Email`
- injects both into request context

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/memberships/model.go)

Defines:

- membership model
- membership roles
- add-member input

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/memberships/service.go)

Implements the first real authorization rule in the backend:

- only `manager` members can add another choir member

It also requires the actor to already belong to the choir before listing memberships.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/memberships/http.go)

Registers the new membership endpoints:

- `POST /api/v1/choirs/{choirID}/memberships`
- `GET /api/v1/choirs/{choirID}/memberships`

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/choirs/service.go)

Choir service now expects actor identity for:

- create
- get
- list

This makes choir access membership-aware rather than tenant-wide.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/choirs/http.go)

Choir handlers now read:

- tenant from request context
- actor from request context

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/users/service.go)

Adds actor resolution by email inside a tenant.

### [choirs_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/choirs_repository.go)

Now does two important Stage 3 things:

- creates a choir inside a transaction
- inserts the creator as the first `manager` membership

This keeps the first write flow membership-aware from its first successful commit.

### [memberships_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/memberships_repository.go)

Implements membership persistence and lookup with explicit SQL.

### [users_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/users_repository.go)

Adds tenant-scoped user lookup by email for actor resolution.

### [repositories_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/repositories_integration_test.go)

Now covers Stage 3 behavior such as:

- choir creation with initial manager membership
- member-scoped choir listing
- membership creation and listing

## New Protected Request Contract

Actor-protected choir and membership routes now require:

```text
X-Tenant-Slug: coral-jovem-asa-norte
X-User-Email: ana@example.com
```

## What This Slice Still Does Not Do

This stage does not yet implement:

- real JWT or identity provider integration
- tenant-wide admin roles
- update/delete choir flows
- richer membership lifecycle like deactivation
- event- or file-level authorization

Those remain later work.
