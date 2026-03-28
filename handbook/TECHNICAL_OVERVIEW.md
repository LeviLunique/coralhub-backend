# Technical Overview

This document explains how CoralHub Backend is structured and why it is structured that way.

## Product Context

CoralHub is a multi-tenant choir management platform.
The first tenant is `Coral Jovem Asa Norte`, but the backend is designed for additional tenants from the start.

## Architectural Style

The application uses a pragmatic modular monolith.

That means:

- code is organized by business capability
- handlers are thin
- services hold application logic
- SQL is explicit
- interfaces are used only at real boundaries

The project intentionally avoids:

- ORM-based persistence
- generic repositories
- microservices
- architecture layers added only for ceremony

## Runtime Shape

The repository produces two binaries:

- `cmd/api`
- `cmd/worker`

The API serves HTTP requests.
The worker processes scheduled notifications and queue cleanup.

Both binaries share the same modules and data model.

Main repository areas:

- `cmd/` for process entrypoints
- `internal/modules/` for business capabilities
- `internal/platform/` for shared operational code
- `internal/store/postgres/` for concrete PostgreSQL implementations
- `internal/integrations/` for storage and push adapters
- `db/migrations/` and `db/queries/` for explicit SQL

## Business Modules

Current modules include:

- `tenants` for tenant bootstrap and tenant metadata lookup
- `choirs` for choir creation and retrieval
- `users` for tenant-scoped user records
- `memberships` for choir membership and role enforcement
- `voicekits` for rehearsal asset grouping
- `files` for file metadata and storage orchestration
- `events` for scheduling choir events and reminders
- `notifications` for worker queue processing
- `devices` for mobile push token persistence
- `audit` for selective tenant-safe change history

## Multi-Tenant Rules

Tenant safety is a first-class requirement.

Core rules:

- tenant-owned rows carry `tenant_id`
- reads and writes are tenant-scoped in SQL
- tenant-local uniqueness includes `tenant_id`
- storage object keys include tenant context
- audit rows carry tenant context

The backend uses a shared-database, shared-schema model.

## Authorization Model

Protected requests are evaluated against:

- authenticated or resolved user identity
- current tenant context
- tenant-scoped role or membership

The current baseline still uses lightweight header-based actor resolution in some flows, but the architecture is designed around explicit tenant and actor context rather than anonymous access.

## Persistence Strategy

The persistence model is SQL-first:

- schema lives in `db/migrations/`
- queries live in `db/queries/`
- `sqlc` generates low-level query bindings
- PostgreSQL repositories map between generated rows and module models

This keeps data access explicit and avoids hidden ORM behavior.

## Storage and Notification Boundaries

Real external boundaries are handled through focused interfaces:

- S3-compatible object storage for file binaries
- FCM for push delivery

The files module stores metadata in PostgreSQL and binaries in object storage.
The notifications module claims scheduled rows from PostgreSQL and delegates delivery to the push integration.

## Core Workflows

### Tenant Bootstrap

Public flows can look up tenant branding and bootstrap data by slug without exposing privileged data.

### Choir and Membership Management

Choirs and users are tenant-scoped.
Memberships establish who can view choir data and who can act as a manager.

### Voice Kits and Files

Voice kits group rehearsal assets by choir.
Managers can upload supported files, while visibility is constrained by membership.

### Events and Reminders

Event writes are transactional.
Reminder rows are generated during event creation and updates so scheduling stays consistent with the event record.

### Notification Processing

The worker claims due reminders with `FOR UPDATE SKIP LOCKED`, sends notifications, retries transient failures, and marks terminal outcomes explicitly.

### Audit History

Critical business transitions write tenant-scoped audit rows rather than adopting full event sourcing.

## Reference Decisions

For the decision rationale behind this overview, see:

- [Application Architecture ADR](adr/application-architecture.md)
- [Infrastructure ADR](adr/infrastructure.md)
- [Multi-Tenant Data Model ADR](adr/multi-tenant-data-model.md)
- [Auth and Tenant Resolution ADR](adr/auth-and-tenant-resolution.md)
