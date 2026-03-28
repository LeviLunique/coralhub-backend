# CoralHub Architecture Summary

This document is the short architectural overview for the CoralHub backend.

Use it when you want the system shape quickly before diving into the full ADR set and implementation guide.

## 1. What This Backend Is

CoralHub backend is:

- a multi-tenant backend for choir management
- a pragmatic modular monolith
- implemented in Go
- split into one API process and one worker process

The initial tenant is:

- `Coral Jovem Asa Norte`

That tenant is seed data, not hardcoded product identity.

## 2. Core Stack

- HTTP router: `chi`
- Database: `PostgreSQL`
- DB access: `pgx + sqlc`
- File storage: S3-compatible storage
- Local object storage: `MinIO`
- Push notifications: `FCM`
- Background processing: PostgreSQL-backed worker with polling and `FOR UPDATE SKIP LOCKED`
- CI: `GitHub Actions`

## 3. Architectural Shape

The backend uses:

- `cmd/api` for the API process
- `cmd/worker` for the background worker
- `internal/modules` for business capabilities
- `internal/platform` for cross-cutting concerns
- `internal/store/postgres` for concrete persistence
- `internal/integrations` for external adapters

The code is organized by business capability, not by artificial architecture layers.

## 4. What The Architecture Intentionally Avoids

The project does not use:

- ORM-based persistence
- generic repositories
- Clean Architecture purism
- Hexagonal Architecture everywhere
- one-use-case-per-file ceremony
- microservices

## 5. Multi-Tenant Rules

The backend is multi-tenant from the start.

This means:

- tenant-owned rows carry `tenant_id`
- tenant-aware authorization is required
- tenant context must be explicit in protected flows
- files, notifications, and audit history stay tenant-scoped

## 6. Operational Model

Production direction from the docs is:

- AWS
- ECS Fargate
- RDS PostgreSQL
- private S3
- Secrets Manager
- CloudWatch
- OpenTelemetry-compatible observability

The API and worker remain separate services in production.

## 7. Where To Read Next

For the canonical source of truth, continue with:

1. [AI_IMPLEMENTATION_GUIDE.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/AI_IMPLEMENTATION_GUIDE.md)
2. [0001-architecture.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0001-architecture.md)
3. [0003-multi-tenant-data-model.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0003-multi-tenant-data-model.md)
4. [0004-auth-and-tenant-resolution.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/adr/0004-auth-and-tenant-resolution.md)
5. [INFRASTRUCTURE_BLUEPRINT.md](/Users/levilunique/Workspace/Go/coralhub-backend/docs/INFRASTRUCTURE_BLUEPRINT.md)
