# Notifications and Delivery

This guide explains the notification worker and FCM delivery integration.

## What It Does

This feature area adds:

- due notification claiming with `FOR UPDATE SKIP LOCKED`
- lease-aware worker processing
- retry scheduling
- invalid-token handling
- FCM-backed delivery when Firebase is enabled

## How It Works

Flow:

1. event workflows create pending scheduled notifications
2. the worker claims due rows from PostgreSQL
3. the worker sends notifications through the configured sender
4. transient failures are retried
5. invalid tokens are deactivated
6. terminal rows are marked explicitly

The worker can still run locally with Firebase disabled.

## Why It Matters

This is the main asynchronous workflow in the backend.
It turns reminder scheduling into real delivery behavior without adding an external queue service.

## How To Verify

Start dependencies:

```bash
make compose-up
```

Start the worker with Firebase disabled:

```bash
make run-worker
```

Expected result:

- the worker starts successfully
- local development can use the no-op sender path

If you already have due notifications, the worker should process them after a poll cycle.

For real FCM verification:

1. set `FIREBASE_ENABLED=true`
2. provide `FIREBASE_CREDENTIALS_FILE`
3. ensure a user has an active device token
4. create or force a due scheduled notification
5. run the worker again

Expected result:

- valid tokens receive delivery attempts
- invalid tokens are deactivated

Automated validation:

```bash
make test
make vet
make build
```
