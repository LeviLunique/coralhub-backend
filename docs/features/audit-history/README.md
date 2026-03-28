# Audit History

This document explains the Stage 9 slice implemented on branch:

- `feat/audit-history`

Commit:

- `35645c3`

## Goal

Add selective tenant-safe audit history for the critical flows already implemented in the backend.

This slice adds:

- tenant-scoped `audit_log`
- audit writes for membership add
- audit writes for event create, update, and cancel
- audit writes for notification generated, sent, failed, and invalid-token transitions
- basic repository-level audit retrieval for support and verification

## Important Scope Choice

Stage 9 in the roadmap says audit retrieval is needed only if the product needs it.

This slice therefore implements retrieval at the repository layer:

- `ListByTenantID`

It does not add a public or protected HTTP endpoint yet.
That keeps the change small and avoids inventing a product-facing audit API before there is a concrete product requirement for one.

## Important Deferred Item

The implementation guide also suggests storing history for file upload and remove actions.

This slice intentionally does **not** audit file upload and remove yet.
The reason is pragmatic:

- current file upload first stores in object storage, then persists metadata
- current file delete first removes from object storage, then deactivates metadata

Adding reliable audit history to those flows without widening the slice would require reworking the current storage and metadata consistency strategy.
Stage 9 therefore focuses on the flows where the backend already has clean transactional boundaries:

- memberships
- events
- notifications

## Runtime Flow

1. a write flow reaches a transactional PostgreSQL repository
2. the business row changes are applied
3. an audit row is written in the same transaction
4. the transaction commits, so business state and history stay aligned

For notification delivery transitions:

1. the worker claims a notification
2. the repository marks it `sent`, `failed`, or `invalid_token`
3. the repository writes the audit row in the same transaction

## File-By-File Explanation

### [000008_init_audit_log.up.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/migrations/000008_init_audit_log.up.sql)

Adds `audit_log` with:

- `tenant_id`
- `entity_type`
- `entity_id`
- `action`
- `actor_id`
- `occurred_at`
- `payload_json`

It also adds the Stage 9 indexes for:

- recent tenant history
- per-entity tenant history

### [audit.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/audit.sql)

Defines explicit SQL for:

- creating audit rows
- listing recent tenant audit rows

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/audit/model.go)

Defines the shared audit model plus the action constants used by the audited slices.

The actions are intentionally pragmatic, not generic diff tracking.

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/audit/repository.go)

Defines the small audit repository contract:

- create
- list by tenant

### [audit_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/audit_repository.go)

Implements the concrete PostgreSQL audit repository on top of `sqlc`.

It also exposes the shared `createAuditLog` helper used by the other PostgreSQL repositories so audit writes stay explicit and consistent.

### [memberships_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/memberships_repository.go)

Membership creation now runs inside a transaction and writes:

- `membership.added`

This audit row includes the acting manager in `actor_id`.

### [events_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/events_repository.go)

Event writes now also create audit rows:

- `event.created`
- `event.updated`
- `event.canceled`
- `notification.generated`

The event repository keeps audit writes inside the same transaction as event and reminder changes.

### [notifications_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/notifications_repository.go)

Notification finalization now writes:

- `notification.sent`
- `notification.failed`
- `notification.invalid_token`

These audit rows are written in the same lease-safe transaction that updates notification status.

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/repository.go)

Event repository input now carries `ActorUserID` for create, update, and cancel so the audit rows can capture the acting manager.

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/memberships/repository.go)

Membership creation now carries `ActorUserID` so the audit row records who added the member.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/memberships/service.go)

Passes the acting manager ID through to the repository.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/service.go)

Passes actor context through the repository write methods so Stage 9 can remain actor-aware without moving audit logic into handlers.

### [repositories_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/repositories_integration_test.go)

Extends integration coverage so the audited repositories prove:

- membership add writes an audit row
- event create, update, cancel, and reminder generation write audit rows
- notification sent, failed, and invalid-token transitions write audit rows
- tenant-scoped audit listing works

## What This Slice Does Not Yet Do

This stage still does not implement:

- file upload/remove audit
- membership removal audit
- an audit HTTP endpoint
- cross-entity support search or filtering

Those can be added later when there is a concrete product need.
