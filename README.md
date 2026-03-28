# CoralHub Backend

Backend for CoralHub, a multi-tenant choir management platform built as a pragmatic modular monolith in Go.

The platform starts with `Coral Jovem Asa Norte` as the initial tenant seed. That tenant is data, not hardcoded product identity.

The documentation entrypoint is [docs/INDEX.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/INDEX.md). A compact architecture overview lives in [docs/ARCHITECTURE_SUMMARY.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/ARCHITECTURE_SUMMARY.md).

## Stack

- Go
- chi
- PostgreSQL
- pgx
- sqlc
- MinIO for local S3-compatible storage

## Quick Start

1. Copy `.env.example` to `.env`.
2. Start local dependencies with `make compose-up`.
3. Follow [LOCAL_ENV_SETUP.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/LOCAL_ENV_SETUP.md) if you need to generate local secrets or align Docker credentials.
4. Apply migrations with your preferred migration runner. The files use `golang-migrate` naming.
5. Start the API with `make run-api`.
6. Start the worker with `make run-worker`.
7. Verify the stack with [LOCAL_SMOKE_TEST.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/LOCAL_SMOKE_TEST.md).

The local PostgreSQL container is exposed on host port `5433` to avoid collisions with existing PostgreSQL installations bound to `5432`.
The app reads explicit DB settings from `.env`: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `DB_SSL_MODE`.

## Useful Commands

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
- `make compose-up`
- `make compose-down`

## Repository Shape

The codebase follows the documented modular monolith structure:

- `cmd/` for process entrypoints
- `internal/platform/` for cross-cutting bootstrap code
- `internal/modules/` for business capabilities
- `internal/store/postgres/` for concrete PostgreSQL wiring
- `db/migrations/` and `db/queries/` for explicit SQL

## Architecture Summary

CoralHub backend is a pragmatic modular monolith with:

- `cmd/api` and `cmd/worker` as separate processes
- business capabilities under `internal/modules`
- explicit PostgreSQL queries under `db/queries`
- `pgx + sqlc` for persistence
- MinIO locally and S3-compatible storage in production
- tenant-aware authorization and data ownership from day one

The deeper architectural documents are:

- [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/AI_IMPLEMENTATION_GUIDE.md)
- [0001-architecture.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0001-architecture.md)
- [0003-multi-tenant-data-model.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0003-multi-tenant-data-model.md)
- [0004-auth-and-tenant-resolution.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0004-auth-and-tenant-resolution.md)

## Initial Database Seed

The first migration creates the tenant bootstrap tables and seeds:

- slug: `coral-jovem-asa-norte`
- display name: `Coral Jovem Asa Norte`
