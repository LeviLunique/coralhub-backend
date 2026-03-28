# Feature Implementation Order

This file tracks the backend implementation order from the roadmap in `docs/AI_IMPLEMENTATION_GUIDE.md`.

It is intended to answer:

- which roadmap stages already have delivered work
- which stages are only partially started
- which feature slices still remain before the backend is complete
- which branch carried an implemented slice
- which short-lived branch name is a reasonable candidate for an upcoming slice

This file is therefore both:

- a history of completed slices
- a forward-looking implementation tracker

For planned stages, the branch names below are recommendations only.
They are not pre-created long-lived branches.

## 1. Current Stage Status

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

## 2. Completed Slices

### Stage 0. Repository Bootstrap

- Status: completed
- Branch: `main`
- Commit: `ef3a243`
- Summary:
  - created the repository scaffold
  - added `go.mod`, `Makefile`, CI placeholder, Dockerfile, Compose file, and local env template
  - added API and worker entrypoints
  - added config, logger, HTTP bootstrap, DB bootstrap
  - created the initial tenant schema and seeded `Coral Jovem Asa Norte`

### Stage 1. Platform Bootstrap

- Status: completed
- Branch: `main`
- Commit: `e9db203`
- Summary:
  - completed the local development bootstrap expected by Stage 1
  - replaced single `DATABASE_URL` with explicit `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `DB_SSL_MODE`
  - aligned `docker-compose.yml` with `.env`
  - aligned PostgreSQL and MinIO local credentials with the `.env` file
  - documented local environment setup and smoke-test flow

### Stage 2 Foundation. Public Tenant Bootstrap Endpoint

- Status: completed as a Stage 2 foundation slice
- Branch: `feat/tenant-bootstrap`
- Commit: `85581ee`
- Summary:
  - added a tenant module
  - added a public endpoint to retrieve tenant bootstrap information by slug
  - added explicit SQL for tenant bootstrap retrieval
  - added service and repository layers for tenant bootstrap
  - wired the endpoint into the API router
  - added tests for service behavior and router behavior

### Stage 2. First Vertical Slice: Choirs and Users

- Status: completed
- Branch: `feat/choirs-users`
- Commit: `c0b6b00`
- Summary:
  - added `choirs` and `users` schema with tenant-aware uniqueness rules
  - added explicit SQL and generated `sqlc` queries for choir and user create/get/list flows
  - added choir and user modules with thin HTTP handlers and focused services
  - added tenant context middleware for protected Stage 2 routes using `X-Tenant-Slug`
  - wired choir and user routes into the API router
  - added PostgreSQL-backed repository integration tests

## 3. Why Stage 2 Started With A Read Flow

The original roadmap lists `choirs` and `users` as the first Stage 2 slice.

The current implementation started Stage 2 with a public tenant bootstrap read flow first because:

- ADR 0004 says protected operations must not trust arbitrary tenant input
- the auth baseline was not implemented yet
- a public tenant bootstrap endpoint is explicitly compatible with the tenant resolution guidance
- this validated a real module, real HTTP routing, and explicit SQL without introducing unsafe write semantics too early

This was a deliberate sequencing choice, not a change in architecture direction.

## 4. Why Stage 2 Pulled In Tenant Context Middleware

The implementation guide recommends `/api/v1/choirs` style routes, but ADR 0004 also says privileged operations should not trust arbitrary tenant data forever.

To keep Stage 2 small while still validating tenant-owned module flows, this slice introduced:

- minimal tenant context middleware using `X-Tenant-Slug`

This is a narrow sequencing pull-forward, not a final auth solution.
Stage 3 still needs to harden this into authenticated actor plus tenant context handling.

## 5. Remaining Roadmap To Backend Completion

### Stage 3. Memberships And Authorization Baseline

- Status: completed
- Branch: `feat/memberships-auth`
- Commit: `d433c15`
- Summary:
  - added `choir_members` schema and explicit membership SQL
  - added actor request context resolved from tenant plus tenant-scoped user email
  - made choir create transactional and auto-created the first manager membership
  - made choir reads membership-aware instead of tenant-wide
  - added membership endpoints with first manager-role authorization checks

### Stage 4. Voice Kits and File Metadata

- Status: completed
- Branch: `feat/voice-kits-file-metadata`
- Commit: `fc3d738`
- Summary:
  - added `voice_kits` and `kit_files` schema with tenant-scoped constraints and indexes
  - added a `voicekits` module for create, get, list, and delete flows
  - added a `files` module for file metadata create, list, and delete flows
  - enforced member visibility and manager-only write/delete rules using the Stage 3 membership baseline
  - added PostgreSQL-backed integration tests for voice kit and file metadata persistence

### Stage 5. S3/MinIO File Upload Integration

- Status: completed
- Branch: `feat/file-storage`
- Commit: `8e6e1c9`
- Summary:
  - added a concrete S3-compatible storage adapter using the official AWS SDK
  - changed file creation from JSON metadata input to multipart upload orchestration
  - generated tenant-aware storage keys and stored metadata only after upload success
  - added pre-signed download URL generation
  - added MinIO-backed integration coverage for upload, download URL, and delete behavior

### Stage 6. Events and Reminder Scheduling

- Status: completed
- Branch: `feat/events-reminders`
- Commit: `6816aba`
- Summary:
  - added `events` and `scheduled_notifications` schema with the Stage 6 indexes and constraints
  - added an `events` module for create, get, list, update, and cancel flows
  - enforced manager-only event writes and membership-aware event reads
  - added reminder policy rules for `day_before` and `hour_before`
  - made event create, update, and cancel flows transactional so reminder scheduling stays atomic

### Stage 7. Notification Worker

- Status: completed
- Branch: `feat/notification-worker`
- Commit: `b1f3600`
- Summary:
  - expanded `scheduled_notifications` with attempts, lease, sent timestamp, and last error fields
  - added explicit queue queries for due-job claiming and worker state transitions
  - implemented a notifications module with retry and invalid-delivery classification rules
  - implemented a real worker polling loop with `FOR UPDATE SKIP LOCKED` and lease-aware completion
  - intentionally deferred device-token deactivation because the repo still has no device token slice

### Stage 8. FCM Integration

- Status: completed
- Branch: `feat/fcm-integration`
- Commit: `e214459`
- Summary:
  - added a concrete FCM sender under `internal/integrations/push/fcm`
  - added minimal `device_tokens` persistence for token lookup and deactivation
  - classified FCM errors into transient vs invalid-token outcomes
  - integrated the worker with the real FCM sender when Firebase is enabled
  - kept a deliberate local-development fallback to the no-op sender when Firebase is disabled

### Stage 9. Audit History

- Status: completed
- Branch: `feat/audit-history`
- Commit: `35645c3`
- Summary:
  - added `audit_log` schema with tenant-scoped indexes
  - added explicit SQL and a concrete PostgreSQL audit repository with tenant-scoped retrieval
  - added audit writes for membership add, event create/update/cancel, reminder generation, and notification sent/failed/invalid-token transitions
  - kept audit writes transactional with the business state transitions they describe
  - intentionally deferred file upload/remove audit and an audit HTTP endpoint to keep the Stage 9 slice small and coherent

### Stage 10. Hardening

- Status: partially completed
- Branch: `feat/hardening`
- Commit: `afc1aac`
- Summary:
  - added structured JSON error responses with stable error codes and request IDs
  - hardened JSON decoding to reject unknown fields and multiple payloads
  - made HTTP handler timeout configurable
  - added a basic `/metrics` endpoint with HTTP, worker, notification, and storage counters
  - added notification retention cleanup in the worker for terminal notification rows
  - intentionally deferred broad pagination so the hardening slice stayed small and coherent

### Stage 11. CI Finalization and Documentation

- Status: completed locally
- Branch: `chore/ci-finalization`
- Commit: `4358895`
- Summary:
  - expanded CI to include `staticcheck`, `golangci-lint`, `govulncheck`, `Trivy`, and `gitleaks`
  - added local parity targets in `Makefile` for CI-relevant checks
  - kept `.env.example` aligned with the latest runtime configuration
  - improved `README.md` so local run and validation docs are easy to find
  - added `docs/ARCHITECTURE_SUMMARY.md` and updated `docs/INDEX.md` for faster onboarding
  - documentation is now visible to Git alongside the CI and repo entrypoint files and can be committed normally

## 6. Current Implemented Feature Branches

- `main`
  - repository bootstrap
  - platform bootstrap
- `develop`
  - currently points to the same commit as `main`
  - note: this is not part of the documented trunk-based recommendation
- `feat/tenant-bootstrap`
  - public tenant bootstrap endpoint
- `feat/choirs-users`
  - Stage 2 choirs and users slice
- `feat/memberships-auth`
  - Stage 3 memberships and authorization baseline
- `feat/voice-kits-file-metadata`
  - Stage 4 voice kits and file metadata
- `feat/file-storage`
  - Stage 5 S3/MinIO file upload integration
- `feat/events-reminders`
  - Stage 6 events and reminder scheduling
- `feat/notification-worker`
  - Stage 7 notification worker
- `feat/fcm-integration`
  - Stage 8 FCM integration
- `feat/audit-history`
  - Stage 9 audit history
- `chore/ci-finalization`
  - Stage 11 CI finalization and documentation entrypoints

## 7. Recommended Next Feature

The next strong candidate is:

- no remaining roadmap stage after Stage 11; the next work should be follow-up slices such as the deferred Stage 10 pagination work

That slice should likely include:

- the remaining deferred Stage 10 pagination work
- any production deployment workflow additions beyond baseline CI
- timeout handling
- metrics
- cleanup and retention rules

## 8. Maintenance Rule For This File

After each meaningful vertical slice:

- update the stage status if progress changed
- add the completed slice under the appropriate stage
- record the real branch and commit
- keep planned future stages in place until they are implemented

This file should remain complete even before the backend is complete.
That means it must track both delivered work and the remaining roadmap.
