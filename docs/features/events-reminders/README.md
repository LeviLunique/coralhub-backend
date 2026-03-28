# Events and Reminder Scheduling

This document explains the Stage 6 slice implemented on branch:

- `feat/events-reminders`

Commit:

- `feat(stage6): add events and reminder scheduling`

## Goal

Add choir events plus PostgreSQL-backed reminder scheduling, without implementing the worker delivery loop yet.

This slice adds:

- tenant-scoped `events`
- tenant-scoped `scheduled_notifications`
- manager-only create, update, and cancel event flows
- member-visible event read and list flows
- reminder generation rules for active choir members
- transactional reminder replacement when events change

## Runtime Flow

For event creation:

1. actor context is resolved from `X-Tenant-Slug` and `X-User-Email`
2. the HTTP handler decodes the event payload
3. the service checks that the actor is a manager of the target choir
4. the service loads active choir memberships
5. reminder times are generated with the event policy
6. the repository creates the event and inserts scheduled reminders in one transaction

For event update:

1. the actor resolves the existing event through membership-scoped access
2. the service checks manager role on the event choir
3. the policy rebuilds the reminder schedule
4. the repository updates the event, cancels pending reminders, and inserts the new ones in one transaction

For event cancel:

1. the actor resolves the event through membership-scoped access
2. the service checks manager role
3. the repository deactivates the event and marks pending reminders as `canceled` in one transaction

## File-By-File Explanation

### [000005_init_events_and_scheduled_notifications.up.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/migrations/000005_init_events_and_scheduled_notifications.up.sql)

Adds:

- `events`
- `scheduled_notifications`

The tables are tenant-scoped from the start and include the indexes the roadmap calls for:

- `events (choir_id, start_at)`
- `scheduled_notifications (status, scheduled_for)`
- `scheduled_notifications (event_id)`

The scheduled notifications table also has a partial unique index so the same pending reminder identity cannot be inserted twice.

### [events.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/events.sql)

Defines the explicit event SQL for:

- create
- update
- get by membership
- list by choir
- cancel

It also includes a query for active choir member user IDs, although the Stage 6 service currently derives reminder recipients from the memberships module.

### [scheduled_notifications.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/scheduled_notifications.sql)

Defines explicit SQL for:

- create scheduled notification
- cancel pending notifications for an event
- list notifications for verification and future worker use

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/model.go)

Defines the Stage 6 event and reminder models plus the current allowed constants:

- event types: `rehearsal`, `presentation`, `other`
- reminder types: `day_before`, `hour_before`
- notification statuses: `pending`, `canceled`

### [repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/repository.go)

Defines the small repository contract the service needs:

- create
- update
- get
- list
- cancel

It deliberately keeps reminder persistence coupled to event writes because Stage 6 requires atomic scheduling behavior.

### [policy.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/policy.go)

Implements the current reminder policy:

- one reminder 24 hours before the event
- one reminder 1 hour before the event

If either timestamp is already in the past at scheduling time, that reminder is skipped.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/service.go)

This is the Stage 6 business layer.

It handles:

- payload normalization
- event type validation
- start time validation
- manager authorization for writes
- membership-aware visibility for reads
- reminder schedule generation from choir membership

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/http.go)

Adds the HTTP routes:

- `POST /api/v1/choirs/{choirID}/events`
- `GET /api/v1/choirs/{choirID}/events`
- `GET /api/v1/events/{eventID}`
- `PUT /api/v1/events/{eventID}`
- `DELETE /api/v1/events/{eventID}`

Handlers stay thin and only map request/response behavior to the service.

### [events_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/events_repository.go)

Implements the concrete PostgreSQL repository using explicit `sqlc` queries.

The important Stage 6 behavior is transactional:

- create event + insert reminders
- update event + cancel pending reminders + insert replacement reminders
- cancel event + cancel pending reminders

### [router.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/http/router.go)

Mounts the Stage 6 event routes inside the protected actor-aware route group.

### [main.go](/Users/levilunique/Workspace/Go/coralhub-backend/cmd/api/main.go)

Constructs the event repository and service and injects them into the API router.

### [service_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/service_test.go)

Verifies the Stage 6 business rules:

- reminder schedule generation
- manager-only writes
- reminder build on create
- manager-only cancel

### [repositories_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/repositories_integration_test.go)

Adds repository integration coverage for:

- event create
- reminder insertion
- event update with reminder replacement
- event cancel with reminder cancellation

## What This Slice Does Not Yet Do

This stage intentionally stops at scheduling persistence.

It does not yet implement:

- polling for due reminders
- `FOR UPDATE SKIP LOCKED`
- retry policy
- push delivery
- FCM integration

Those belong to Stages 7 and 8.
