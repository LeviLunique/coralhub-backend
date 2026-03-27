# CoralHub Backend

Backend for CoralHub, a multi-tenant choir management platform built as a pragmatic modular monolith in Go.

The platform starts with `Coral Jovem Asa Norte` as the initial tenant seed. That tenant is data, not hardcoded product identity.

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
3. Apply migrations with your preferred migration runner. The files use `golang-migrate` naming.
4. Start the API with `make run-api`.
5. Start the worker with `make run-worker`.

The local PostgreSQL container is exposed on host port `5433` to avoid collisions with existing PostgreSQL installations bound to `5432`.

## Useful Commands

- `make fmt`
- `make vet`
- `make test`
- `make build`
- `make sqlc`
- `make compose-up`
- `make compose-down`

## Repository Shape

The codebase follows the documented modular monolith structure:

- `cmd/` for process entrypoints
- `internal/platform/` for cross-cutting bootstrap code
- `internal/modules/` for business capabilities
- `internal/store/postgres/` for concrete PostgreSQL wiring
- `db/migrations/` and `db/queries/` for explicit SQL

## Initial Database Seed

The first migration creates the tenant bootstrap tables and seeds:

- slug: `coral-jovem-asa-norte`
- display name: `Coral Jovem Asa Norte`
