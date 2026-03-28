# CoralHub Backend

CoralHub Backend is a multi-tenant choir management backend built in Go.

The platform starts with `Coral Jovem Asa Norte` as the initial tenant seed. That tenant is data, not hardcoded product identity.

## Documentation

The remote-friendly documentation entrypoint is [handbook/README.md](handbook/README.md).

Recommended reading order:

1. [handbook/README.md](handbook/README.md)
2. [handbook/LOCAL_DEVELOPMENT.md](handbook/LOCAL_DEVELOPMENT.md)
3. [handbook/TECHNICAL_OVERVIEW.md](handbook/TECHNICAL_OVERVIEW.md)
4. [handbook/INFRASTRUCTURE.md](handbook/INFRASTRUCTURE.md)
5. [handbook/FEATURES.md](handbook/FEATURES.md)

Detailed decision records and implementation history remain in:

- [`docs/adr/`](docs/adr/)
- [`docs/features/`](docs/features/)

## Stack

- Go
- chi
- PostgreSQL
- pgx
- sqlc
- S3-compatible object storage
- MinIO for local object storage
- FCM for push notifications

## Quick Start

1. Copy `.env.example` to `.env`.
2. Read [handbook/LOCAL_DEVELOPMENT.md](handbook/LOCAL_DEVELOPMENT.md).
3. Start local dependencies with `make compose-up`.
4. Apply database migrations with your preferred migration runner.
5. Start the API with `make run-api`.
6. Start the worker with `make run-worker`.
7. Verify the service with `curl http://127.0.0.1:8080/healthz`.

The local PostgreSQL container uses host port `5433` to avoid conflicts with local installations on `5432`.

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
- `make run-api`
- `make run-worker`

## Repository Structure

- `cmd/` contains the API and worker entrypoints.
- `internal/modules/` contains business capabilities such as tenants, choirs, memberships, events, files, notifications, and audit history.
- `internal/platform/` contains configuration, routing, observability, logging, and request context helpers.
- `internal/store/postgres/` contains the concrete PostgreSQL implementations.
- `internal/integrations/` contains external adapters such as S3-compatible storage and FCM.
- `db/migrations/` and `db/queries/` contain the explicit SQL that defines and accesses the data model.

## Product Model

CoralHub is a shared backend for multiple tenants.

Core rules:

- tenant-owned data carries `tenant_id`
- authorization is tenant-aware
- storage keys include tenant context
- API and worker run as separate processes
- SQL is explicit and first-class
- the system does not use an ORM or generic repositories
