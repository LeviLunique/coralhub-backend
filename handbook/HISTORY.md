# Delivery History

This document summarizes the roadmap progress that has already been delivered in the repository.

## Stage Status

- Stage 0: completed
- Stage 1: completed
- Stage 2: completed
- Stage 3: completed
- Stage 4: completed
- Stage 5: completed
- Stage 6: completed
- Stage 7: completed
- Stage 8: completed
- Stage 9: completed
- Stage 10: partially completed
- Stage 11: completed

## Completed Slices

### Stage 0 and Stage 1

Delivered:

- repository bootstrap
- API and worker entrypoints
- local PostgreSQL and MinIO setup
- initial tenant bootstrap schema and seed
- `.env.example`, Compose support, and local run commands

### Stage 2

Delivered:

- public tenant bootstrap lookup
- tenant-scoped choirs and users flows
- initial protected route handling with tenant context

Details:

- [Tenant bootstrap](features/tenant-bootstrap.md)
- [Choirs and users](features/choirs-and-users.md)

### Stage 3

Delivered:

- choir memberships
- manager and member roles
- actor-aware authorization baseline

Details:

- [Memberships and authorization](features/memberships-and-authorization.md)

### Stage 4 and Stage 5

Delivered:

- voice kit metadata
- file metadata
- S3-compatible upload flow
- download URLs

Details:

- [Voice kits and files](features/voice-kits-and-files.md)

### Stage 6

Delivered:

- events
- reminder scheduling
- transactional reminder regeneration during event changes

Details:

- [Events and reminders](features/events-and-reminders.md)

### Stage 7 and Stage 8

Delivered:

- PostgreSQL-backed notification worker
- retry and lease handling
- FCM-backed delivery
- invalid token deactivation

Details:

- [Notifications and delivery](features/notifications-and-delivery.md)

### Stage 9

Delivered:

- tenant-scoped audit history for critical business transitions

Details:

- [Audit history](features/audit-history.md)

### Stage 10

Delivered:

- structured error responses
- strict JSON decoding
- configurable handler timeout
- metrics endpoint
- notification retention cleanup

Still deferred:

- broad pagination across list endpoints

Details:

- [Hardening](features/hardening.md)

### Stage 11

Delivered:

- expanded CI baseline
- local parity commands
- documentation entrypoints

Details:

- [Quality and CI](features/quality-and-ci.md)

## Main Follow-Up Work

The most visible remaining work after the documented roadmap is:

- broad pagination
- deployment workflows beyond CI
- release automation
- richer production operations
