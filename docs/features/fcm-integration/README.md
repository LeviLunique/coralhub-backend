# FCM Integration

This document explains the Stage 8 slice implemented on branch:

- `feat/fcm-integration`

Commit:

- `feat(stage8): add fcm integration`

## Goal

Replace the Stage 7 no-op notification sender with a real FCM-backed delivery path.

This slice adds:

- a minimal `device_tokens` store
- a concrete FCM sender
- per-token error classification
- token deactivation on invalid token responses
- worker integration with Firebase-aware configuration

## Important Sequencing Note

The implementation guide lists a fuller `devices` slice with HTTP endpoints later.

This Stage 8 implementation does **not** pull that entire slice forward.

Instead it adds only the minimum needed for push delivery to become functional:

- token persistence
- active-token lookup
- token deactivation

That keeps the change small and lets the worker send real push notifications without introducing the full devices API yet.

## Deliberate Local-Development Deviation

The worker now supports two startup paths:

- real FCM sender when `FIREBASE_ENABLED=true`
- Stage 7 no-op sender when Firebase is disabled

This is a deliberate deviation from a hard fail-always posture.

Reason:

- without a Firebase service account JSON, every local worker run would fail to start
- the repo still needs to remain testable and runnable for contributors who do not have Firebase credentials locally

When Firebase is enabled, the real push delivery path is used.

## Runtime Flow

1. Stage 6 creates pending scheduled notifications
2. Stage 7 claims due notifications in the worker
3. the Stage 8 FCM sender loads active device tokens for the notification user
4. it sends a multicast FCM message to those tokens
5. invalid tokens are deactivated immediately
6. the sender returns one high-level delivery result back to the notifications service:
   - `sent`
   - `transient_failure`
   - `invalid_token`
7. the Stage 7 queue logic updates notification state accordingly

## File-By-File Explanation

### [000007_init_device_tokens.up.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/migrations/000007_init_device_tokens.up.sql)

Adds the minimal token table needed for FCM delivery:

- `tenant_id`
- `user_id`
- `platform`
- `token`
- `active`

The table includes:

- a deduplicating unique constraint on `(tenant_id, token)`
- the early index recommended by the implementation guide on `(user_id, active)`

### [devices.sql](/Users/levilunique/Workspace/Go/coralhub-backend/db/queries/devices.sql)

Defines explicit SQL for:

- create or reactivate a device token
- list active tokens for a user
- deactivate a token

### [model.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/modules/devices/model.go)

Defines the minimal device token model needed for Stage 8.

This is intentionally small and does not yet add a full devices service or HTTP layer.

### [devices_repository.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/devices_repository.go)

Implements PostgreSQL token persistence using explicit `sqlc` queries.

The create path reactivates an existing token if the same token is registered again for the tenant.

### [sender.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/integrations/push/fcm/sender.go)

This is the core Stage 8 change.

It:

- builds the Firebase app and messaging client
- loads active tokens for the notification user
- sends a multicast message through FCM
- classifies per-token errors as:
  - transient
  - invalid token
- deactivates invalid tokens

### [sender_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/integrations/push/fcm/sender_test.go)

Adds focused tests for:

- no active tokens
- partial success with invalid token cleanup
- transient failures when retryable tokens remain
- all-invalid-token handling

### [main.go](/Users/levilunique/Workspace/Go/coralhub-backend/cmd/worker/main.go)

The worker now:

- builds the device token repository
- creates the real FCM sender when Firebase is enabled
- falls back to the no-op sender when Firebase is disabled

### [config.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/platform/config/config.go)

Adds the `firebase` config group expected by the implementation guide:

- `FIREBASE_ENABLED`
- `FIREBASE_CREDENTIALS_FILE`

It fails fast when Firebase is enabled but the credentials file is missing.

### [repositories_integration_test.go](/Users/levilunique/Workspace/Go/coralhub-backend/internal/store/postgres/repositories_integration_test.go)

Adds PostgreSQL integration coverage for:

- device token create/reactivation
- active token listing
- token deactivation

## What This Slice Does Not Yet Do

This stage still does not implement:

- device registration HTTP endpoints
- device listing HTTP endpoints
- user notification read APIs
- richer localized notification content

Those remain later product work.
