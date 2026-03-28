# Features and Roadmap Status

This document summarizes the current product capabilities delivered in the backend.

## Current Status

The roadmap is complete through Stage 11, with one notable follow-up area still deferred from Stage 10:

- broad pagination across list endpoints

Everything else has a working baseline in the repository.

## Platform Foundation

Delivered:

- repository bootstrap
- API and worker entrypoints
- configuration loading
- PostgreSQL and MinIO local development support
- initial tenant seed for `Coral Jovem Asa Norte`

## Tenant Bootstrap

Delivered capability:

- public tenant bootstrap lookup by slug

Representative endpoint:

- `GET /api/v1/public/tenants/{tenantSlug}`

Purpose:

- support branded bootstrap and tenant discovery flows

Detailed documentation:

- [Detailed guide](features/tenant-bootstrap.md)

## Choirs and Users

Delivered capability:

- tenant-scoped choir creation and reads
- tenant-scoped user creation and reads

Representative endpoints:

- `POST /api/v1/choirs`
- `GET /api/v1/choirs`
- `GET /api/v1/choirs/{choirID}`
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{userID}`

Detailed documentation:

- [Detailed guide](features/choirs-and-users.md)

## Memberships and Authorization Baseline

Delivered capability:

- choir membership records
- manager and member roles
- actor-aware authorization baseline
- membership-aware choir visibility

Representative behavior:

- first choir creator becomes the first manager
- choir membership management requires manager access

Detailed documentation:

- [Detailed guide](features/memberships-and-authorization.md)

## Voice Kits and Rehearsal Files

Delivered capability:

- voice kit CRUD baseline
- file metadata handling
- real S3-compatible uploads
- pre-signed download URLs
- membership-aware visibility
- manager-only write and delete flows

Representative endpoints:

- `POST /api/v1/choirs/{choirID}/voice-kits`
- `GET /api/v1/choirs/{choirID}/voice-kits`
- `GET /api/v1/voice-kits/{voiceKitID}`
- `POST /api/v1/voice-kits/{voiceKitID}/files`
- `GET /api/v1/files/{fileID}/download-url`

Detailed documentation:

- [Detailed guide](features/voice-kits-and-files.md)

## Events and Reminders

Delivered capability:

- choir event creation, update, listing, and cancelation
- reminder scheduling during event writes
- member-visible reads
- manager-only writes

Reminder policy currently supports:

- `day_before`
- `hour_before`

Detailed documentation:

- [Detailed guide](features/events-and-reminders.md)

## Notification Delivery

Delivered capability:

- PostgreSQL-backed notification queue
- worker claiming with `FOR UPDATE SKIP LOCKED`
- retry handling and lease ownership
- FCM integration
- invalid-token deactivation

Operational behavior:

- transient failures are retried
- invalid device tokens are deactivated
- terminal outcomes are retained for a configurable time before cleanup

Detailed documentation:

- [Detailed guide](features/notifications-and-delivery.md)

## Audit History

Delivered capability:

- tenant-scoped audit log table
- audit rows for membership changes
- audit rows for event lifecycle changes
- audit rows for notification generation and delivery outcomes

The backend currently stores audit history at the repository layer without exposing a public audit API.

Detailed documentation:

- [Detailed guide](features/audit-history.md)

## Hardening and Observability

Delivered capability:

- structured JSON error envelopes
- strict JSON decoding
- configurable handler timeout
- `/metrics` endpoint
- notification retention cleanup

Detailed documentation:

- [Detailed guide](features/hardening.md)

## CI and Delivery Tooling

Delivered capability:

- GitHub Actions CI for quality and security checks
- local parity commands in `Makefile`
- `.env.example` aligned with current runtime configuration

Key local command:

- `make ci`

Detailed documentation:

- [Detailed guide](features/quality-and-ci.md)

## Remaining Follow-Up Work

The main documented follow-up areas are:

- broad pagination across list endpoints
- deployment workflows beyond CI
- release automation and richer production operations

## Detailed Slice History

If you need the per-stage branch and commit history, see:

- [Delivery History](HISTORY.md)
