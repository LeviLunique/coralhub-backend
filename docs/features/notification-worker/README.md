# Notification Worker

This document explains the Stage 7 slice implemented on branch:

- `feat/notification-worker`

Commit:

- `feat(stage7): add notification worker`

## Goal

Process due scheduled notifications safely in a worker instead of sending anything inline during event creation.

This slice adds:

- worker-oriented notification queries
- `FOR UPDATE SKIP LOCKED` claiming
- lease-based processing ownership
- retry scheduling with a max-attempt policy
- invalid-delivery classification
- a real worker loop in `cmd/worker`

## Important Sequencing Note

The repository still does not have:

- a `device_tokens` module
- an FCM adapter

So this slice intentionally stops at queue processing behavior.

The worker currently uses a temporary no-op sender that marks claimed notifications as delivered. That keeps Stage 7 focused on:

- durable queue behavior
- locking
- retry transitions
- worker orchestration

It does **not** try to pull Stage 8 forward.

Because there is no device-token registry yet, invalid-token handling is implemented as final notification classification:

- `invalid_token`

Actual token deactivation remains deferred to the later devices and FCM slices.

## Runtime Flow

1. event creation from Stage 6 inserts pending rows in `scheduled_notifications`
2. the worker claims due rows with `FOR UPDATE SKIP LOCKED`
3. claimed rows move to `processing` and receive a lease timestamp
4. the notifications service invokes the sender
5. the repository marks each notification as:
   - `sent`
   - `pending` again with a retry time
   - `failed`
   - `invalid_token`
6. row updates require the same lease timestamp, so stale workers cannot overwrite a newer claim

## File-By-File Explanation

### [000006_alter_scheduled_notifications_for_worker.up.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/migrations/000006_alter_scheduled_notifications_for_worker.up.sql)

Extends `scheduled_notifications` with the worker state it needs:

- `attempts`
- `last_error`
- `processing_started_at`
- `sent_at`

It also expands the status constraint to include:

- `invalid_token`

### [notifications.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/notifications.sql)

Adds the explicit worker queries for:

- claiming due notifications with `FOR UPDATE SKIP LOCKED`
- marking notifications sent
- retrying transient failures
- marking permanent failure
- marking invalid token

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/notifications/model.go)

Defines the Stage 7 notification worker model and statuses.

It also defines the delivery result kinds used by the worker service:

- `sent`
- `transient_failure`
- `invalid_token`

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/notifications/repository.go)

Defines the focused repository contract the worker logic needs:

- claim due
- mark sent
- retry
- mark failed
- mark invalid token

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/notifications/service.go)

This is the Stage 7 business logic.

It:

- claims due work
- applies retry rules
- stops retrying at the configured max attempts
- keeps lease ownership explicit through `processing_started_at`
- treats lease loss as a safe no-op so overlapping workers do not corrupt state

### [worker.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/notifications/worker.go)

Adds the polling loop used by the worker process.

It:

- runs one cycle immediately at startup
- keeps polling on the configured interval
- logs when notifications are processed

### [notifications_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/notifications_repository.go)

Implements the concrete PostgreSQL queue behavior with explicit `sqlc` queries.

The important detail is lease-aware state transitions:

- completion updates require the same `processing_started_at`
- if the lease changed, the repository returns a lease-lost error and the stale worker does not win

### [main.go](/Users/levilunique/Workspace/Go/coralhub-backend/cmd/worker/main.go)

Wires the worker process to:

- PostgreSQL
- the notifications repository
- the notifications service
- the worker loop

For Stage 7 it uses a no-op sender that reports successful delivery.

### [service_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/notifications/service_test.go)

Adds focused unit coverage for:

- sent notifications
- transient failure retry scheduling
- permanent failure after max attempts
- invalid token classification
- unknown sender result rejection

### [repositories_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/repositories_integration_test.go)

Adds PostgreSQL integration coverage for:

- due-queue claiming
- sent transition
- retry transition
- invalid-token transition
- failed transition

## What This Slice Does Not Yet Do

This stage still does not implement:

- real FCM delivery
- device token persistence
- token deactivation in a `device_tokens` table
- user notification listing APIs

Those belong to later slices.
