# Hardening

This document explains the Stage 10 hardening slice implemented on branch:

- `feat/hardening`

Commit:

- `afc1aac`

## Goal

Improve the operational credibility of the backend without redesigning the existing feature modules.

This slice adds:

- structured JSON error responses with stable error codes and request IDs
- stricter JSON request decoding
- configurable HTTP handler timeout
- a basic metrics endpoint
- notification retention cleanup in the worker

## Important Sequencing Decision

Stage 10 in `docs/AI_IMPLEMENTATION_GUIDE.md` also lists pagination.

I intentionally did **not** implement broad pagination in this slice.
Doing it correctly across every existing list endpoint would have widened the change substantially and cut against the request to keep the slice small, coherent, and testable.

So this is a real Stage 10 hardening slice, but not the entire remaining Stage 10 backlog.

## Runtime Changes

### Structured Errors

All current HTTP handlers now return a consistent error envelope:

```json
{
  "error": {
    "code": "tenant_header_required",
    "message": "X-Tenant-Slug header is required",
    "request_id": "..."
  }
}
```

### Request Validation Hardening

JSON handlers now reject:

- empty JSON bodies
- malformed JSON
- unknown JSON fields
- multiple JSON documents in one body

That keeps the handlers thin while making input behavior more predictable.

### Timeout Handling

The handler timeout is now configurable through:

- `HTTP_HANDLER_TIMEOUT`

This replaces the previous fixed timeout value in the router bootstrap.

### Metrics

The API now exposes:

- `GET /metrics`

The metrics are intentionally basic and in-process.
They currently track:

- HTTP request count
- HTTP request duration sum and count
- worker poll count
- worker processed-notification count
- notification delivery result count
- notification cleanup deleted count
- storage upload failure count

### Notification Retention Cleanup

The worker now deletes old terminal notification rows based on:

- `WORKER_NOTIFICATION_RETENTION`

Terminal states cleaned up by this slice are:

- `sent`
- `failed`
- `canceled`
- `invalid_token`

This keeps the queue table from growing forever after delivery outcomes are final.

## File-By-File Explanation

### [.env.example](/Users/levilunique/Workspace/Go/coralhub-backend/.env.example)

Adds the new hardening configuration:

- `HTTP_HANDLER_TIMEOUT`
- `WORKER_NOTIFICATION_RETENTION`

### [config.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/config/config.go)

Loads the new timeout and retention settings with sensible defaults.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/web/http.go)

Adds shared HTTP helpers for:

- JSON responses
- structured error responses
- strict JSON decoding

This package is intentionally below the router package so feature handlers can reuse it without introducing an import cycle.

### [middleware.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/http/middleware.go)

Now:

- logs status consistently
- emits HTTP request metrics
- uses the shared structured error helper for tenant and actor middleware failures
- provides the configurable timeout middleware

### [router.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/http/router.go)

Now:

- accepts the configured handler timeout
- exposes `/metrics`

### [metrics.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/observability/metrics.go)

Implements a small in-process metrics registry and renders a Prometheus-style text response.

### [notifications.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/notifications.sql)

Adds explicit SQL for deleting old terminal notifications.

### [notifications_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/notifications_repository.go)

Adds the concrete retention cleanup query method.

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/notifications/service.go)

Adds the service-level cleanup operation and emits notification delivery metrics when final delivery-state transitions succeed.

### [worker.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/notifications/worker.go)

Now:

- increments worker poll metrics
- increments processed-count metrics
- runs retention cleanup each cycle
- logs cleanup deletions when rows are removed

### [service.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/service.go)

Now increments the storage upload failure metric when the object storage upload itself fails.

### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/choirs/http.go)
### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/users/http.go)
### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/memberships/http.go)
### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/voicekits/http.go)
### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/files/http.go)
### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/events/http.go)
### [http.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/tenants/http.go)

These handlers now use the shared structured error and strict JSON helpers instead of each module carrying its own ad hoc JSON/error helpers.

## What This Slice Does Not Yet Do

This slice still does not implement:

- broad pagination across the existing list endpoints
- richer timeout response shaping beyond context deadlines
- tracing or OpenTelemetry instrumentation

Those remain valid follow-up hardening work.
