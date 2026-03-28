# Contributing

This guide explains how to make safe changes to CoralHub Backend without needing outside repository context.

## Recommended Reading Order

Before changing code, read:

1. [TECHNICAL_OVERVIEW.md](TECHNICAL_OVERVIEW.md)
2. [adr/README.md](adr/README.md)
3. [FEATURES.md](FEATURES.md)
4. the relevant detailed guide under [`features/`](features/)

If your change affects runtime behavior or deployment assumptions, also read:

- [INFRASTRUCTURE.md](INFRASTRUCTURE.md)
- [PRODUCTION_BLUEPRINT.md](PRODUCTION_BLUEPRINT.md)

## Project Rules

CoralHub Backend uses:

- Go
- `chi`
- PostgreSQL
- `pgx + sqlc`
- S3-compatible storage
- FCM

Core engineering rules:

- keep handlers thin
- keep SQL explicit
- avoid ORM-based persistence
- avoid generic repositories
- organize code by business capability
- implement coherent vertical slices
- keep tenant isolation explicit in data and authorization

## How To Approach a Change

Use this workflow:

1. understand which business module owns the change
2. identify whether the change affects schema, SQL, service logic, HTTP behavior, worker behavior, or docs
3. make the smallest coherent change that completes the slice
4. run the relevant validation
5. update handbook pages if behavior or usage changed

## Where Code Belongs

Use these placement rules:

- `internal/modules/` for business logic by capability
- `internal/platform/` for shared operational code
- `internal/store/postgres/` for concrete PostgreSQL repositories
- `internal/integrations/` for external service adapters
- `db/migrations/` for schema changes
- `db/queries/` for explicit SQL consumed by `sqlc`

Do not introduce new architecture layers unless they solve a real problem already present in the codebase.

## Common Change Patterns

### Adding a new database-backed behavior

Typical flow:

1. add or update a migration
2. add explicit SQL in `db/queries/`
3. regenerate `sqlc`
4. update the PostgreSQL repository
5. update the service
6. update the handler or worker if needed
7. add or update tests

### Changing a protected API flow

Check:

- how tenant context is resolved
- how actor identity is resolved
- where role or membership is enforced
- whether the SQL is tenant-scoped

### Changing file or notification behavior

Remember that these are external-boundary features.
Keep SDK-specific details inside integrations or focused interfaces, not inside handlers.

## Tenant Safety Checklist

Before merging a feature that touches tenant-owned data, confirm:

1. every tenant-owned row carries `tenant_id`
2. reads filter by `tenant_id`
3. writes filter by `tenant_id` where applicable
4. uniqueness rules include `tenant_id` where needed
5. tests cover wrong-tenant or missing-tenant behavior when relevant

If any of these are missing, the feature is not complete.

## HTTP and API Expectations

Current project conventions:

- public flows are clearly separated from protected flows
- protected flows use explicit tenant and actor context
- handlers return structured JSON errors
- health endpoints exist at `/healthz` and `/api/v1/healthz`

For staged local development, some protected examples still use:

- `X-Tenant-Slug`
- `X-User-Email`

Treat those as current repository behavior, not as an excuse to weaken tenant-aware authorization.

## Validation Expectations

Minimum validation for most changes:

```bash
make test
make vet
make build
```

Preferred full validation before opening a PR:

```bash
make ci
```

If you changed SQL or schema:

```bash
make sqlc
git diff --exit-code
```

## Documentation Expectations

Update the handbook when you change:

- contributor workflow
- runtime behavior
- production assumptions
- feature behavior visible to other engineers

Update the relevant detailed feature guide when you change a documented feature area.

## When To Stop and Reassess

Pause and rethink the change if you find yourself adding:

- a new abstraction that mostly forwards calls
- a repository interface with no real boundary value
- a new package structure that duplicates existing module ownership
- tenant-blind shortcuts
- infrastructure assumptions that conflict with the AWS target

In those cases, the design is probably drifting away from the documented architecture.
